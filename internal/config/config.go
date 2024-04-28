package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	// Configure rclone
	var once sync.Once
	once.Do(func() {
		rc_configfile.Install()
		// Silence, rclone!
		conf := rc_fs.GetConfig(ctx)
		conf.LogLevel = rc_fs.LogLevelWarning
		conf.Progress = true
		conf.MultiThreadSet = true
	})

	if remote == "" {
		if unattended {
			return "", fmt.Errorf("no remote specified")
		} else {
			logger.Warn("You did not specify a remote destination.")
			return chooseRemote(), nil
		}
	}

	// If remote is an absolute path, use local backend
	if filepath.IsAbs(remote) {
		absPath, err := filepath.Abs(remote)
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	// Remote was specified by user
	definedRemotes := rc_config.FileSections()
	noRemotes := len(definedRemotes) == 0

	// Check if specified remote is defined in rclone's config
	isDefined := func(rem string) bool {
		logger.Debugf("Validating remote: %s", rem)
		for i, r := range definedRemotes {
			logger.Debugf("%d: %s", i+1, r)
			if r == rem {
				return true
			}
		}
		return false
	}

	// If specified remote exists, return it
	if isDefined(remote) {
		return remote, nil
	}

	// At this point, the specified remote doesn't exist
	if unattended {
		// Nobody is there to choose, can only use local backend
		if noRemotes {
			// User has no remotes
			logger.Debug("No remotes defined: using local backend")
		} else {
			// User has remotes, but can't choose
			logger.Infof("The specified remote doesn't exist: %s", remote)
			ShowRcloneConfigPath()
		}
		return "", nil
	} else {
		// User is there to choose
		if noRemotes {
			// But has no remotes
			logger.Info("You don't have any remote configured.")
			if b := booleanChoice("Would you like to create one now?", unattended); b {
				name := rc_config.NewRemoteName()
				err := rc_config.NewRemote(ctx, name)
				if err != nil {
					return "", err
				}
				return name, nil
			} else {
				logger.Info("Then we can only use the local backend.")
				return "", nil
			}
		} else {
			// User has defined other remotes, ask to choose one
			logger.Warnf("The specified remote doesn't exist: %s", remote)
			return chooseRemote(), nil
		}
	}
}

func chooseRemote() string {
	var c string
	for c == "" {
		c = rc_config.ChooseRemote()
	}
	return c
}

func booleanChoice(question string, unattended bool) bool {
	if unattended {
		logger.Info("No user no choice lol")
		return false
	} else {
		for {
			logger.Info(question)
			logger.Info("name> ")
			answer := strings.ToLower(rc_config.ReadLine())

			switch {
			case answer == "y" || answer == "yes":
				return true
			case answer == "n" || answer == "no":
				return false
			default:
				continue
			}
		}
	}
}

func ShowRcloneConfigPath() {
	if configPath := rc_config.GetConfigPath(); configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			logger.Debug("The configuration file for rclone doesn't exist, but it will use this path:")
		} else {
			logger.Debug("Check rclone's configuration at:")
		}
		logger.Debugf("%s\n", configPath)
	} else {
		// logger.Info("Configuration is in memory only")
	}
}
