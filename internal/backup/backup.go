package backup

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/0x07cf-dev/go-backup/internal/lang"
	"github.com/0x07cf-dev/go-backup/internal/logger"
	"github.com/0x07cf-dev/go-backup/internal/notify"
	_ "github.com/rclone/rclone/backend/local"
	_ "github.com/rclone/rclone/backend/webdav"
)

type BackupSession struct {
	Config   *MachineConfig
	Opts     *BackupOpts
	Notifier *notify.Notifier
	// Internals
	context   context.Context
	processed map[string]bool
	mu        sync.Mutex
}

type BackupOpts struct {
	Remote         string
	RemoteRoot     string
	Uploading      bool
	Simulate       bool
	Interactive    bool
	Debug          bool
	EnvFilePath    string
	ConfigFilePath string
	LangFilePath   string
	Languages      []string
}

type BackupOptFunc func(*BackupOpts)

func defaultBackupOpts() *BackupOpts {
	// Determine system language
	langs := []string{}
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}
	p := strings.Split(lang, "_")
	if len(p) > 0 && len(p[0]) > 0 {
		logger.Infof("Detected system language: '%s' (%s)", p[0], lang)
		langs = append(langs, p[0])
	}
	langs = append(langs, "en")

	return &BackupOpts{
		Remote:         "",
		RemoteRoot:     "Backups",
		Uploading:      true,
		Simulate:       false,
		Interactive:    true,
		Debug:          false,
		ConfigFilePath: "configs/config.json",
		LangFilePath:   "lang.en.toml",
		Languages:      langs,
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
		opts.Interactive = interactive
	}
}

func WithDebug(debug bool) BackupOptFunc {
	return func(opts *BackupOpts) {
		opts.Debug = debug
	}
}

func WithEnvFile(path string) BackupOptFunc {
	return func(opts *BackupOpts) {
		if path != "" {
			opts.EnvFilePath = path
		}
	}
}

func WithConfigFile(path string) BackupOptFunc {
	return func(opts *BackupOpts) {
		if path != "" {
			opts.ConfigFilePath = path
		}
	}
}

func WithLangFile(path string) BackupOptFunc {
	return func(opts *BackupOpts) {
		if path != "" {
			opts.LangFilePath = path
		}
	}
}

func WithLanguage(langs ...string) BackupOptFunc {
	return func(opts *BackupOpts) {
		opts.Languages = langs
	}
}

func NewSession(options ...BackupOptFunc) *BackupSession {
	ctx := context.Background()
	opts := defaultBackupOpts()
	for _, fn := range options {
		fn(opts)
	}

	// Load environment file
	if err := loadEnvFile(opts.EnvFilePath); err != nil {
		logger.Error("There was an error loading environment variables.")
	}

	// Load rclone config (if user didn't specify remote, it will be picked from it)
	loadRemoteConfig(ctx, opts)

	// Load language
	lang.LoadLanguage(opts.LangFilePath, opts.Languages...)

	// Load current machine parameters from configuration
	machineConfig, err := getMachineConfig(opts.ConfigFilePath)
	if err != nil {
		if machineConfig == nil {
			logger.Fatal(err)
		}
	}

	// Load notifier parameters from environment
	notifier, err := notify.NewNotifierFromEnv()
	if err != nil {
		logger.Errorf("Error creating notifier: %s", err.Error())
		logger.Error("The backup will still be performed, but notifications will not be sent.")
	}

	return &BackupSession{
		Config:    machineConfig,
		Opts:      opts,
		Notifier:  notifier,
		context:   ctx,
		processed: make(map[string]bool),
	}
}

func (session *BackupSession) Backup() {
	t0 := time.Now()
	wg := sync.WaitGroup{}

	numPaths := len(session.Config.Paths)
	numPreCmds := len(session.Config.PreCommands)
	numPostCmds := len(session.Config.PostCommands)

	if numPaths == 0 && numPreCmds == 0 && numPostCmds == 0 {
		logger.Warn("Nothing to do. Please take a look at the configuration file.")
		return
	}

	// Ternary operator is sometimes useful :(
	var sb strings.Builder
	sb.WriteString("Initializing ")
	if session.Opts.Uploading {
		sb.WriteString("upload ")
	} else {
		sb.WriteString("download ")
	}
	if session.Opts.Simulate {
		sb.WriteString("simulation...")
	} else {
		sb.WriteString("session...")
	}
	sb.WriteString(fmt.Sprintf("(%s)", session.Opts.Remote))
	logger.Info(sb.String())

	session.Heartbeat("start", false)

	// Execute pre commands
	preErrCh := make(chan BackupError, numPreCmds)
	if numPreCmds > 0 {
		logger.Info("Executing pre-transfer commands...")
		executeCmds(preErrCh, session.Config.PreCommands, session.Config.CmdOutput)
	}
	close(preErrCh)

	// Spawn transfer goroutines
	transferErrCh := make(chan BackupError, numPaths)
	if session.Opts.Uploading {
		for _, path := range session.Config.Paths {
			wg.Add(1)
			go session.uploadPath(path, &wg, transferErrCh, session.Opts.Simulate)
		}
	} else {
		for _, path := range session.Config.Paths {
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
		executeCmds(postErrCh, session.Config.PostCommands, session.Config.CmdOutput)
	}
	close(postErrCh)

	// Notify status to user
	status, statusEmoji := getStatus(transferErrCh, preErrCh, postErrCh, session.Opts.Languages...)
	session.NotifyStatus(status, statusEmoji, "package")

	// Ping healthchecks
	session.Heartbeat("", true)
	logger.Infof("ALL DONE! Time taken: %v\n", time.Since(t0))
}

func (session *BackupSession) Heartbeat(endpoint string, withLog bool) {
	if session.Opts.Interactive {
		// Session is interactive, no heartbeats
		return
	}
	if session.Notifier != nil {
		resp, err := session.Notifier.SendHeartbeats(endpoint, withLog)
		if err != nil {
			logger.Errorf("Error sending heartbeat: %s", err)
		}
		logger.Infof("Heartbeat Status: '%s'", resp)
	}
}

func (session *BackupSession) NotifyStatus(status string, statusTags ...string) {
	if session.Notifier != nil {
		msgTitle := fmt.Sprintf(
			"%s - %s",
			strings.ToUpper(session.Config.Hostname),
			strings.ToTitle(session.Notifier.Topic),
		)
		msgTags := make([]string, 0, len(statusTags)+2)
		msgTags = append(msgTags, statusTags...)
		msgTags = append(msgTags, session.Config.Hostname, session.Notifier.Topic)

		resp, err := session.Notifier.Send(msgTitle, status, msgTags)
		if err != nil {
			logger.Errorf("Error sending status notification: %s", err)
		} else {
			logger.Infof("Notifier Status: '%s'", resp)
		}
	}
}
