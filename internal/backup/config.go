package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/0x07cf-dev/go-backup/internal/logger"
	"github.com/joho/godotenv"
	rc_config "github.com/rclone/rclone/fs/config"
	rc_configfile "github.com/rclone/rclone/fs/config/configfile"
	"golang.org/x/sys/windows/registry"
)

type GlobalConfig struct {
	Machines []MachineConfig `json:"machines"`
}

type MachineConfig struct {
	Hostname     string   `json:"hostname"`
	Paths        []string `json:"paths"`
	CmdOutput    bool     `json:"output"`
	PreCommands  []string `json:"pre"`
	PostCommands []string `json:"post"`
}

func defaultMachineConfig() *MachineConfig {
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("Error retrieving hostname:", err)
		hostname = "Default"
	}

	return &MachineConfig{
		Hostname:     hostname,
		Paths:        []string{},
		CmdOutput:    true,
		PreCommands:  []string{},
		PostCommands: []string{},
	}
}

func loadEnvFile(path string) error {
	logger.Info("Loading environment...")
	if err := godotenv.Load(path); err != nil {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("Error loading .env file: '%v'", err)
		}
		exeDir := filepath.Dir(exe)
		if err := godotenv.Load(filepath.Join(exeDir, ".env")); err != nil {
			return fmt.Errorf("Error loading .env file: '%v'", err)
		}
	}
	return nil
}

func loadRemoteConfig(ctx context.Context, opts *BackupOpts) {
	rc_configfile.Install()
	rc_config.ShowConfigLocation()

	availableRemotes := rc_config.FileSections()
	for i, r := range availableRemotes {
		logger.Debugf("%d: %s", i+1, r)
	}

	// No user-specified remote or no remotes to choose from
	if opts.Remote == "" || len(rc_config.FileSections()) == 0 {
		chooseRemote := func(ctx context.Context, interactive bool) (string, error) {
			if len(availableRemotes) > 0 {
				// We have remotes but there is no user to pick one
				if !interactive {
					opts.Remote = availableRemotes[0]
				}
				logger.Info("No remote specified. You need to choose one:")
				return rc_config.ChooseRemote(), nil
			} else {
				// We don't have remotes and there is no user to create one
				if !interactive {
					return "", fmt.Errorf("no remote exists")
				}
				logger.Info("No remote exists. You need to create one:")
				name := rc_config.NewRemoteName()
				err := rc_config.NewRemote(ctx, name)
				if err != nil {
					return "", err
				}
				return name, nil
			}
		}

		remote, err := chooseRemote(ctx, opts.Interactive)
		if err != nil {
			logger.Fatalf("The program cannot operate without a remote: %s", err.Error())
		}
		opts.Remote = remote
	}
}

func getMachineConfig(configPath string) (*MachineConfig, error) {
	var machineConfig *MachineConfig

	globalConfig, err := readConfigFile(configPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logger.Info("Config file not found. One with default options will be created.")
			machineConfig = defaultMachineConfig()
			writeConfigFile(configPath, &GlobalConfig{
				Machines: []MachineConfig{*machineConfig},
			})
			return machineConfig, err
		} else {
			return nil, err
		}
	}

	logger.Info("Found config file, reading...")
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	for _, machine := range globalConfig.Machines {
		if machine.Hostname == hostname {
			machineConfig = &machine
			break
		}
	}

	modified := false
	if machineConfig == nil {
		logger.Info("Current machine is not configured.")
		machineConfig = defaultMachineConfig()
		modified = true
	}

	// Manipulate paths before use
	for i, p := range machineConfig.Paths {
		expanded, err := cleanPath(p)
		if err != nil {
			return nil, err
		}

		if p != expanded {
			machineConfig.Paths[i] = expanded
			// modified = true
		}
	}

	if modified {
		if err = updateConfigFile(configPath, machineConfig); err != nil {
			return nil, err
		}
	}

	return machineConfig, nil
}

func cleanPath(p string) (string, error) {
	// Expand environment variables (both $UNIX and %WINDOWS%)
	res, err := registry.ExpandString(p)
	if err != nil {
		return "", err
	}

	// Clean
	res = path.Clean(res)
	return res, nil
}

func readConfigFile(path string) (*GlobalConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config GlobalConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	return &config, nil
}

func writeConfigFile(path string, config *GlobalConfig) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	} else {
		logger.Debugf("Created config directory: %s", dir)
	}

	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}

	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logger.Infof("Config file saved at '%s'.", path)
	return nil
}

func updateConfigFile(path string, new *MachineConfig) error {
	logger.Info("Updating current config...")
	globalConfig, err := readConfigFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading existing configuration: %w", err)
	}

	// Config exists, check if this machine is declared
	machineFound := false
	for i, m := range globalConfig.Machines {
		// If the hostname exists, overwrite its config with the new one
		if m.Hostname == new.Hostname {
			// Check if the configuration has changed
			if !reflect.DeepEqual(m, *new) {
				globalConfig.Machines[i] = *new
				if err := writeConfigFile(path, globalConfig); err != nil {
					return fmt.Errorf("error writing updated configuration: %w", err)
				}
			}
			machineFound = true
			break
		}
	}

	// If the hostname does not exist, append its new config
	if !machineFound {
		globalConfig.Machines = append(globalConfig.Machines, *new)
		if err := writeConfigFile(path, globalConfig); err != nil {
			return fmt.Errorf("error writing updated configuration: %w", err)
		}
	}

	return nil
}
