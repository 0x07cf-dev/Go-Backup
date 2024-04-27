package utils

import (
	"os/exec"
	"path"

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

func ParseCommand(c string) (*exec.Cmd, error) {
	// Expand environment variables
	command, err := registry.ExpandString(c)
	if err != nil {
		return nil, err
	}

	systemCmd := exec.Command("cmd.exe", "/C", command)
	return systemCmd, nil
}
