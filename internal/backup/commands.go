package backup

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/0x07cf-dev/go-backup/internal/logger"
	"github.com/0x07cf-dev/go-backup/internal/utils"
)

type CmdFunc func(chan BackupError, string) string

var CommandMap = map[string]CmdFunc{
	"!sleep": cmdSleep,
}

func executeCmds(errCh chan BackupError, commands []string, output bool) {
	for i, command := range commands {
		// If command is in map, execute custom behaviour
		parts := strings.Split(command, " ")
		if len(parts) == 0 {
			errCh <- CmdInvalid.Error(command, "invalid command")
			continue
		}

		cmdName := parts[0]
		if cmdFunc, ok := CommandMap[cmdName]; ok {
			output := cmdFunc(errCh, command)
			logger.Infof("%d° Command Output:\n%s\n", i+1, output)
			continue
		}

		// Otherwise, execute it on the system
		systemCmd, err := utils.ParseCommand(command)
		if err != nil {
			errCh <- CmdInvalid.Error(command, "could not parse command")
			continue
		}

		// Pipe command's output
		stdoutBuf := bytes.Buffer{}
		if output {
			systemCmd.Stdout = &stdoutBuf
		}
		// Errors will be displayed regardless of "output" config variable
		stderrBuf := bytes.Buffer{}
		systemCmd.Stderr = &stderrBuf

		// Run command and display output
		if err := systemCmd.Run(); err != nil {
			logger.Errorf("%d° Command Error: '%s'", i+1, command)
			logger.Error(stderrBuf.String())
			errCh <- CmdFailed.Error(command, stderrBuf.String())
			continue
		} else if output && len(stdoutBuf.String()) > 0 {
			logger.Infof("%d° Command Output:\n%s\n", i+1, stdoutBuf.String())
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
	logger.Infof("Sleeping for %d seconds\n", duration)
	return "Sleeping for " + parts[1] + " seconds"
}
