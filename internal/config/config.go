package config

import (
	"context"
	"fmt"
	"os"
	"sync"

	rc_fs "github.com/rclone/rclone/fs"
	rc_config "github.com/rclone/rclone/fs/config"
	rc_configfile "github.com/rclone/rclone/fs/config/configfile"

	"github.com/0x07cf-dev/go-backup/internal/logger"
	"github.com/0x07cf-dev/go-backup/internal/utils"
	"github.com/spf13/viper"
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
		expanded, err := utils.CleanPath(p)
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

func ValidateRemote(ctx context.Context, remote string, interactive bool) (string, error) {
	var once sync.Once
	once.Do(func() {
		rc_configfile.Install()
		// Silence rclone
		conf := rc_fs.GetConfig(ctx)
		conf.LogLevel = rc_fs.LogLevelWarning
		conf.MultiThreadSet = true
	})

	available := rc_config.FileSections()
	if len(available) == 0 {
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

	// Check if specified remote is defined in rclone's config
	for i, r := range available {
		logger.Debugf("%d: %s", i+1, r)
		if r == remote {
			return r, nil
		}
	}

	// Some remotes are defined, but the specified one isn't
	if !interactive {
		logger.Debug("Session is non-interactive: picking first available remote.")
		return available[0], nil
	} else {
		logger.Infof("The specified remote (%s) doesn't exist. You need to choose one:", remote)
		return rc_config.ChooseRemote(), nil
	}
}
