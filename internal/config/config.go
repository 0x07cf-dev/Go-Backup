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
		logger.Debugf("Found machine in config: %s", hostname)
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

func AsValidRemote(ctx context.Context, remote string, unattended bool) (string, error) {
	var once sync.Once
	once.Do(func() {
		rc_configfile.Install()
		// Silence, rclone!
		conf := rc_fs.GetConfig(ctx)
		conf.LogLevel = rc_fs.LogLevelWarning
		conf.Progress = true
		conf.MultiThreadSet = true
	})

	// If no remotes are configured, summon the wizard
	available := rc_config.FileSections()
	if len(available) == 0 {
		if unattended {
			return "", fmt.Errorf("no remote exists")
		}

		logger.Info("You don't have any remote configured.")
		name := rc_config.NewRemoteName()
		err := rc_config.NewRemote(ctx, name)
		if err != nil {
			return "", err
		}
		return name, nil
	}

	// Check if specified remote is defined in rclone's config
	if remote != "" {
		logger.Debugf("Validating remote: %s", remote)
		for i, r := range available {
			logger.Debugf("%d: %s", i+1, r)
			if r == remote {
				return r, nil
			}
		}
	}

	// Found configured remotes, but specified one is either not among them or null
	if unattended {
		logger.Debug("Session is non-interactive: picking first available remote destination.")
		return available[0], nil
	} else {
		if remote != "" {
			logger.Warnf("The specified remote doesn't exist: %s", remote)
		} else {
			logger.Error("You haven't specified a remote.")
		}
		var chosen string
		for chosen == "" {
			chosen = rc_config.ChooseRemote()
		}
		return chosen, nil
	}
}
