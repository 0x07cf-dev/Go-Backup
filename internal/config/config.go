package config

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	rc_config "github.com/rclone/rclone/fs/config"
	rc_configfile "github.com/rclone/rclone/fs/config/configfile"

	"github.com/0x07cf-dev/go-backup/internal/logger"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows/registry"
)

var Global *GlobalConfig

// GlobalConfig represents the top-level configuration
type GlobalConfig struct {
	Machines []*Machine `json:"machines"`
}

// Machine represents a single machine configuration
type Machine struct {
	Hostname string   `json:"hostname"`
	Paths    []string `json:"paths"`
	Output   bool     `json:"output"`
	Pre      []string `json:"pre"`
	Post     []string `json:"post"`
}

func getConfig() (*GlobalConfig, error) {
	if Global == nil {
		if err := viper.Unmarshal(&Global); err != nil {
			return nil, err
		}
	}
	return Global, nil
}

func GetCurrentMachine() (*Machine, error) {
	var current *Machine

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	globalConfig, err := getConfig()
	if err != nil {
		return nil, err
	}

	// Check if current machine is configured
	for _, m := range globalConfig.Machines {
		if m.Hostname == hostname {
			current = m
			break
		}
	}

	// Check if current machine is configured
	modified := false
	if current != nil {
		logger.Infof("Found machine in config: %s", hostname)
	} else {
		logger.Info("Current machine is not configured.")
		current = &Machine{
			Hostname: hostname,
			Paths:    []string{},
			Output:   true,
			Pre:      []string{},
			Post:     []string{},
		}
		globalConfig.Machines = append(globalConfig.Machines, current)
		//globalConfig.Machines[hostname] = *current
		modified = true
	}

	// Manipulate paths before use
	for i, p := range current.Paths {
		expanded, err := cleanPath(p)
		if err != nil {
			return nil, err
		}

		if p != expanded {
			current.Paths[i] = expanded
			// modified = true
		}
	}

	// Write changes to config
	if modified {
		viper.Set("machines", globalConfig.Machines)
		if err := viper.WriteConfig(); err != nil {
			return nil, err
		}
	}

	return current, nil
}

func cleanPath(p string) (string, error) {
	logger.Debugf("Path before: '%s'", p)

	// Expand environment variables (both $UNIX and %WINDOWS%)
	res, err := registry.ExpandString(p)
	if err != nil {
		return "", err
	}

	// Clean
	res = path.Clean(res)
	logger.Debugf("Path after: '%s'", res)
	return res, nil
}

func ValidateRemote(ctx context.Context, remote *string, interactive bool) {
	var once sync.Once
	once.Do(func() {
		rc_configfile.Install()
		// rc_config.ShowConfigLocation()
	})

	available := rc_config.FileSections()
	found := false

	for i, r := range available {
		logger.Debugf("%d: %s", i+1, r)
		if r == *remote {
			found = true
		}
	}

	// No user-specified remote or no remotes to choose from
	if !found {
		chooseRemote := func() (string, error) {
			if len(available) > 0 {
				if !interactive {
					return available[0], nil
				}

				logger.Info("The remote specified doesn't exist. You need to choose one:")
				return rc_config.ChooseRemote(), nil
			} else {
				if !interactive {
					return "", fmt.Errorf("no remote exists")
				}

				logger.Info("No remotes found. You need to create one:")
				name := rc_config.NewRemoteName()
				err := rc_config.NewRemote(ctx, name)
				if err != nil {
					return "", err
				}
				return name, nil
			}
		}

		rem, err := chooseRemote()
		if err != nil {
			logger.Fatalf("The program cannot operate without a remote: %s", err.Error())
		}
		remote = &rem
	}
}
