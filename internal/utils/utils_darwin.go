package utils

import (
	"os"
	"os/exec"
	"path"
)

func CleanPath(p string) (string, error) {
	// Expand environment variables ($UNIX)
	res := os.ExpandEnv(p)

	// Clean
	res = path.Clean(res)
	return res, nil
}

func ParseCommand(c string) (*exec.Cmd, error) {
	// Expand environment variables
	command := os.ExpandEnv(c)
	systemCmd := exec.Command("bash", "-c", command)
	return systemCmd, nil
}
