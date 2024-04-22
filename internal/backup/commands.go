package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/0x07cf-dev/go-backup/internal/logger"
	"github.com/0x07cf-dev/go-backup/internal/utils"
)

type CmdFunc func(chan BackupError, string) string

var CommandMap = map[string]CmdFunc{
	"!sleep": cmdSleep,
	"cd":     cmdCD,
}

type commandContext struct {
	CWD string
}

var cmdContext commandContext

func executeCmds(errCh chan BackupError, commands []string, output bool) {
	for i, command := range commands {
		parts := strings.Split(command, " ")
		if len(parts) == 0 {
			errCh <- CmdInvalid.Error(command, "invalid command")
			continue
		}

		// If command is in map, execute custom behaviour
		baseCommand := parts[0]
		if cmdFunc, ok := CommandMap[baseCommand]; ok {
			output := cmdFunc(errCh, command)
			logger.Infof("%d° COMMAND OUTPUT (%s):\n%s\n", i+1, command, output)
			continue
		}

		// Otherwise, execute it on the system
		systemCmd, err := utils.ParseCommand(command)
		if err != nil {
			errCh <- CmdInvalid.Error(command, "could not parse command")
			continue
		}

		// Set command working directory
		systemCmd.Dir = cmdContext.CWD

		// Pipe command output
		stdoutBuf := bytes.Buffer{}
		if output {
			systemCmd.Stdout = &stdoutBuf
		}

		// Errors will be displayed regardless of "output" config variable
		stderrBuf := bytes.Buffer{}
		systemCmd.Stderr = &stderrBuf

		// Run command and display output
		if err := systemCmd.Run(); err != nil {
			logger.Errorf("%d° COMMAND FAILURE: '%s'", i+1, command)
			logger.Error(stderrBuf.String())
			errCh <- CmdFailed.Error(command, stderrBuf.String())
			continue
		} else if output && len(stdoutBuf.String()) > 0 {
			logger.Infof("%d° COMMAND OUTPUT (%s):\n%s\n", i+1, command, stdoutBuf.String())
		}
	}

	time.Sleep(1 * time.Second)
}

func cmdSleep(errCh chan BackupError, command string) string {
	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		err := "invalid sleep syntax"
		errCh <- CmdInvalid.Error(command, err)
		return err
	}

	duration, err := strconv.Atoi(parts[1])
	if err != nil {
		err := "invalid sleep duration"
		errCh <- CmdInvalid.Error(command, err)
		return err
	}

	time.Sleep(time.Duration(duration) * time.Second)
	return "Sleeping for " + parts[1] + " seconds"
}

func cmdCD(errCh chan BackupError, command string) string {
	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		err := "invalid cd syntax"
		errCh <- CmdInvalid.Error(command, err)
		return err
	}

	newDir := parts[1]
	absPath, err := filepath.Abs(newDir)
	if err != nil {
		err := "error resolving absolute path for cd"
		errCh <- CmdInvalid.Error(command, err)
		return err
	}

	// Check if the directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		err := "directory does not exist"
		errCh <- CmdInvalid.Error(command, err)
		return err
	}

	cmdContext.CWD = absPath
	return "Changed directory to: " + absPath
}
