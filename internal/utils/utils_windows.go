package utils

import (
	"os"
	"os/exec"
	"path"

	"github.com/0x07cf-dev/go-backup/internal/logger"
	"golang.org/x/sys/windows/registry"
)

func CleanPath(p string) (string, error) {
	// Expand environment variables (%WINDOWS%)
	res, err := registry.ExpandString(p)
	if err != nil {
		return "", err
	}

	// Clean
	res = path.Clean(res)
	return res, nil
}

func ParseCommand(command string) (*exec.Cmd, error) {
	// Expand environment variables
	command, err := registry.ExpandString(command)
	if err != nil {
		return nil, err
	}

	systemCmd := exec.Command("cmd.exe", "/C", command)

	// TODO: fix %VARS% not working
	logger.Debugf("Command: %s || Command Env: %v", systemCmd, systemCmd.Env)
	systemCmd.Env = os.Environ()
	//systemCmd.Env = append(systemCmd.Env, "MY_VAR=some_value")
	return systemCmd, nil
}
