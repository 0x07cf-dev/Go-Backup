package backup

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/0x07cf-dev/go-backup/internal/config"
	"github.com/0x07cf-dev/go-backup/internal/logger"
	"github.com/0x07cf-dev/go-backup/internal/notify"

	//_ "github.com/rclone/rclone/backend/drive"
	//_ "github.com/rclone/rclone/backend/dropbox"
	_ "github.com/rclone/rclone/backend/local"
	//_ "github.com/rclone/rclone/backend/s3"
	_ "github.com/rclone/rclone/backend/webdav"
)

type BackupSession struct {
	Opts     *BackupOpts
	Machine  *config.Machine
	Notifier *notify.Notifier
	// Internals
	context   context.Context
	processed map[string]bool
	mu        sync.Mutex
}

type BackupOpts struct {
	Remote     string
	RemoteRoot string
	Language   string
	Uploading  bool
	Simulate   bool
	Unattended bool
	Debug      bool
}

type BackupOptFunc func(*BackupOpts)

func defaultBackupOpts() *BackupOpts {
	// Determine system language
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}
	p := strings.Split(lang, "_")
	if len(p) > 0 && len(p[0]) > 0 {
		logger.Debugf("Detected system language: '%s' (%s)", p[0], lang)
	} else {
		lang = "en"
	}

	return &BackupOpts{
		Remote:     "default",
		RemoteRoot: "Backups",
		Uploading:  true,
		Simulate:   false,
		Unattended: false,
		Debug:      false,
		Language:   lang,
	}
}

func WithRemote(remote string) BackupOptFunc {
	return func(opts *BackupOpts) {
		if remote != "" {
			opts.Remote = remote
		}
	}
}

func WithRemoteRoot(remoteRoot string) BackupOptFunc {
	return func(opts *BackupOpts) {
		opts.RemoteRoot = remoteRoot
	}
}

func WithDownload() BackupOptFunc {
	return func(opts *BackupOpts) {
		opts.Uploading = false
	}
}

func WithSimulation(simulate bool) BackupOptFunc {
	return func(opts *BackupOpts) {
		opts.Simulate = simulate
	}
}

func WithInteractivity(interactive bool) BackupOptFunc {
	return func(opts *BackupOpts) {
		opts.Unattended = !interactive
	}
}

func WithDebug(debug bool) BackupOptFunc {
	return func(opts *BackupOpts) {
		opts.Debug = debug
	}
}

func WithLanguage(lang string) BackupOptFunc {
	return func(opts *BackupOpts) {
		opts.Language = lang
	}
}

func NewSession(ctx context.Context, options ...BackupOptFunc) *BackupSession {
	opts := defaultBackupOpts()
	for _, fn := range options {
		fn(opts)
	}

	// Load current machine parameters from configuration
	machine, err := config.GetCurrentMachine()
	if err != nil {
		logger.Fatal(err.Error())

	}

	// Load notifier parameters from environment
	notifier, err := notify.NewNotifierFromEnv()
	if err != nil {
		logger.Errorf("Error creating notifier: %s", err.Error())
		logger.Error("The backup will still be performed, but notifications will not be sent.")
	}

	return &BackupSession{
		Opts:      opts,
		Machine:   machine,
		Notifier:  notifier,
		context:   ctx,
		processed: make(map[string]bool),
	}
}

func (session *BackupSession) Backup() {
	t0 := time.Now()
	wg := sync.WaitGroup{}

	numPaths := len(session.Machine.Paths)
	numPreCmds := len(session.Machine.Pre)
	numPostCmds := len(session.Machine.Post)

	if numPaths == 0 && numPreCmds == 0 && numPostCmds == 0 {
		logger.Error("Nothing to do. Please take a look at the configuration file.")
		return
	}

	// Ternary operator is sometimes useful :(
	var sb strings.Builder
	sb.WriteString("INITIALIZING ")
	if session.Opts.Uploading {
		sb.WriteString("UPLOAD ")
	} else {
		sb.WriteString("DOWNLOAD ")
	}
	if session.Opts.Simulate {
		sb.WriteString("SIMULATION ")
	} else {
		sb.WriteString("SESSION ")
	}
	sb.WriteString(fmt.Sprintf("(%s)", session.Opts.Remote))
	logger.Info(sb.String())

	session.Heartbeat("start", false)

	// Execute pre commands
	preErrCh := make(chan BackupError, numPreCmds)
	if numPreCmds > 0 {
		logger.Debug("Executing pre-transfer commands...")
		executeCmds(preErrCh, session.Machine.Pre, session.Machine.Output)
	}
	close(preErrCh)

	// Spawn transfer goroutines
	transferErrCh := make(chan BackupError, numPaths)
	if session.Opts.Uploading {
		logger.Debugf("Starting upload routines... %v", session.Machine.Paths)
		for _, path := range session.Machine.Paths {
			wg.Add(1)
			logger.Debugf("Added routine for path: %s", path)
			go session.uploadPath(path, &wg, transferErrCh, session.Opts.Simulate)
		}
	} else {
		logger.Debugf("Starting download routines... %v", session.Machine.Paths)
		for _, path := range session.Machine.Paths {
			logger.Debugf("Added routine for path: %s", path)
			wg.Add(1)
			go session.downloadPath(path, &wg, transferErrCh, session.Opts.Simulate)
		}
	}

	// Sync goroutines
	wg.Wait()
	close(transferErrCh)

	// Execute post commands
	postErrCh := make(chan BackupError, numPostCmds)
	if numPostCmds > 0 {
		logger.Info("Executing post-transfer commands...")
		executeCmds(postErrCh, session.Machine.Post, session.Machine.Output)
	}
	close(postErrCh)

	// Notify status to user
	status, statusEmoji := getStatus(transferErrCh, preErrCh, postErrCh, session.Opts.Language)
	session.NotifyStatus(status, statusEmoji, "package")

	// Ping healthchecks
	session.Heartbeat("", true)
	logger.Infof("ALL DONE! Time taken: %v\n", time.Since(t0))
}

func (session *BackupSession) Heartbeat(endpoint string, withLog bool) {
	if session.Notifier != nil {
		if !session.Opts.Unattended {
			logger.Debugf("Session is non-interactive: heartbeat will not be sent. %v", session.Notifier.HealthMonitors)
			return
		}
		resp, err := session.Notifier.SendHeartbeats(endpoint, withLog)
		if err != nil {
			logger.Errorf("Error sending heartbeat: %s", err)
		}
		logger.Infof("Heartbeat Status: '%s'", resp)
	} else if session.Opts.Unattended {
		logger.Error("Heartbeats are not configured.")
	}
}

func (session *BackupSession) NotifyStatus(status string, statusTags ...string) {
	if session.Notifier != nil {
		msgTitle := fmt.Sprintf(
			"%s - %s",
			strings.ToUpper(session.Machine.Hostname),
			strings.ToTitle(session.Notifier.Topic),
		)
		msgTags := make([]string, 0, len(statusTags)+2)
		msgTags = append(msgTags, statusTags...)
		msgTags = append(msgTags, session.Machine.Hostname, session.Notifier.Topic)

		resp, err := session.Notifier.Send(msgTitle, status, msgTags)
		if err != nil {
			logger.Errorf("Error sending status notification: %s", err)
		} else {
			logger.Debugf("Notifier Status: '%s'", resp)
			logger.Infof("Post-transfer notification sent.")
		}
	}
}
