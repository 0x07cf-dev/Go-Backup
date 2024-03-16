package backup

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/0x07cf-dev/go-backup/internal/logger"
	rc_fs "github.com/rclone/rclone/fs"
	rc_fspath "github.com/rclone/rclone/fs/fspath"
	rc_ops "github.com/rclone/rclone/fs/operations"
	rc_sync "github.com/rclone/rclone/fs/sync"
)

func initFs(ctx context.Context, path string) (rc_fs.Fs, error) {
	newFs, err := rc_fs.NewFs(ctx, path)
	if err != nil {
		return nil, err
	}

	// Make dir if not found
	if _, err := newFs.List(ctx, ""); errors.Is(err, rc_fs.ErrorDirNotFound) {
		if err = rc_ops.Mkdir(ctx, newFs, ""); err != nil {
			return nil, err
		}
		logger.Infof("%s created!", path)
	}
	return newFs, nil
}

func (session *BackupSession) uploadPath(path string, wg *sync.WaitGroup, errCh chan BackupError, simulate bool) {
	t0 := time.Now()
	defer wg.Done()

	// Mutex lock
	session.mu.Lock()
	defer session.mu.Unlock()
	if _, ok := session.processed[path]; ok {
		logger.Warnf("Path already processed: '%s'", path)
		return
	}
	session.processed[path] = true

	currFile, err := os.Stat(path)
	if err != nil {
		errCh <- PathError.Error(path, err.Error())
		return
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		errCh <- PathError.Error(path, err.Error())
		return
	}
	parent, _, err := rc_fspath.Split(path)
	if err != nil {
		errCh <- PathError.Error(path, err.Error())
		return
	}

	if simulate {
		// totally valid and human errors
		errorChance := 0.2
		if rand.Float32() < float32(errorChance) {
			var syllables []string
			vowels := []rune{'a', 'e', 'i', 'o', 'u'}
			consonants := []rune{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z'}

			for _, c := range consonants {
				for _, v := range vowels {
					syllables = append(syllables, string(c)+string(v))
				}
			}

			randomString := ""
			for i := 0; i < 12; i++ {
				randomString += syllables[rand.Intn(len(syllables))]
				if rand.Intn(3) == 0 {
					randomString += " "
				}
			}

			logger.Infof("Simulating error: '%s'", path)
			simulation := []BackupErrorCode{GenericError, UploadError}
			rnd := (rand.Intn(len(simulation)) * rand.Intn(len(simulation))) % len(simulation)
			errCh <- simulation[rnd].Error(path, randomString)
			return
		}
	}

	// Remote needs to end with ':'
	rem := session.Opts.Remote
	if !strings.HasSuffix(rem, ":") {
		rem += ":"
	}

	// This is a naive approach that simply copies the files/dirs over, overwriting.
	logger.Infof("Uploading: '%s'", path)
	if currFile.IsDir() {
		// Upload directory
		illegal, err := regexp.Compile(`[|<>?\/:*"]`)
		if err != nil {
			errCh <- UploadError.Error(path, err.Error())
			return
		}

		cleanPath := illegal.ReplaceAllString(path, "")
		remoteRoot := rc_fspath.JoinRootPath(
			rem,
			filepath.Join(
				session.Opts.RemoteRoot,
				session.Config.Hostname,
				cleanPath,
			),
		)

		destFs, err := initFs(session.context, remoteRoot)
		if err != nil {
			errCh <- UploadError.Error(path, err.Error())
			return
		}
		srcFs, err := initFs(session.context, absPath)
		if err != nil {
			errCh <- UploadError.Error(path, err.Error())
			return
		}

		// Upload
		if !simulate {
			if err = rc_sync.CopyDir(session.context,
				destFs, // Upload dir destination: remoteRoot/hostname/sourceDirName/...
				srcFs,  // Upload dir source: as defined in config
				true,   // Copy empty dirs?
			); err != nil {
				errCh <- UploadError.Error(path, err.Error())
				return
			}
		} else {
			logger.Infof("Simulated: '%s' --> '%s'", srcFs, destFs)
		}
	} else {
		// Upload file
		remoteRoot := rc_fspath.JoinRootPath(
			rem,
			filepath.Join(
				session.Opts.RemoteRoot,
				session.Config.Hostname,
			),
		)

		destFs, err := initFs(session.context, remoteRoot)
		if err != nil {
			errCh <- UploadError.Error(path, err.Error())
			return
		}
		srcFs, err := initFs(session.context, parent)
		if err != nil {
			errCh <- UploadError.Error(path, err.Error())
			return
		}

		if !simulate {
			if err = rc_ops.CopyFile(
				session.context,
				destFs, // Upload file destination: remoteRoot/hostname/sourceFileName.any
				srcFs,  // Upload file source: as defined in config
				currFile.Name(),
				currFile.Name(),
			); err != nil {
				errCh <- UploadError.Error(path, err.Error())
				return
			}
		} else {
			logger.Infof("Simulated: '%s' --> '%s'", filepath.Join(srcFs.Root(), currFile.Name()), destFs)
		}
	}

	logger.Infof("UPLOADED! (%v)\n", time.Since(t0))
}

// not gonna happen for a while
func (session *BackupSession) downloadPath(path string, wg *sync.WaitGroup, errCh chan BackupError, simulate bool) {
	t0 := time.Now()
	logger.Infof("DOWNLOADED! (%v)\n", time.Since(t0))
}
