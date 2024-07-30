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
	"sleep":  cmdSleep,
	"cd":     cmdCD,
	"export": cmdExport,
}

type CommandOpts struct {
	CWD string
	Env []string
}

var cmdContext CommandOpts

func executeCmds(errCh chan BackupError, commands []string, output bool) {
	// Reset context
	cmdContext.CWD, _ = os.Getwd()
	cmdContext.Env = os.Environ()

	for i, command := range commands {
		ordinal := i + 1
		subCommands := strings.Split(command, "&")
		for _, subCommand := range subCommands {
			subCommand = strings.TrimSpace(subCommand)
			if subCommand == "" {
				continue
			}

			parts := strings.Split(subCommand, " ")
			if len(parts) == 0 {
				errCh <- CmdInvalid.Error(subCommand, "invalid command")
				continue
			}

			outTempl := "%d° (%s): '%s'"
			errTempl := "%d° (%s): '%s'"

			// Parse command and expand environment variables
			systemCmd, err := utils.ParseCommand(subCommand)
			if err != nil {
				errCh <- CmdInvalid.Error(subCommand, "could not parse command")
				continue
			}

			// If command is in map, execute custom behaviour
			// Otherwise, execute it on the system
			baseCommand := parts[0]
			if cmdFunc, ok := CommandMap[baseCommand]; ok {
				output := cmdFunc(errCh, subCommand)
				logger.Infof(outTempl, ordinal, subCommand, output)
				continue
			}

			// Set command working directory
			systemCmd.Dir = cmdContext.CWD
			systemCmd.Env = cmdContext.Env

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
				logger.Errorf(errTempl, ordinal, subCommand)
				logger.Error(stderrBuf.String())
				errCh <- CmdFailed.Error(subCommand, stderrBuf.String())
				continue
			} else if output {
				logger.Infof(outTempl, ordinal, subCommand, stdoutBuf.String())
			}
		}
		time.Sleep(1 * time.Second)
	}
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

func cmdExport(errCh chan BackupError, command string) string {
	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		err := "invalid export syntax"
		errCh <- CmdInvalid.Error(command, err)
		return err
	}

	// Extract key and value from the command
	kv := strings.SplitN(parts[1], "=", 2)
	if len(kv) != 2 {
		err := "invalid export syntax"
		errCh <- CmdInvalid.Error(command, err)
		return err
	}
	k := kv[0]
	v := kv[1]

	cmdContext.Env = append(cmdContext.Env, k+"="+v)
	return "Exported variable: " + k + "=" + v
}
