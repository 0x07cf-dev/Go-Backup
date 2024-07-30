/*
Copyright © 2024 0x07cf-dev <0x07cf@pm.me>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/0x07cf-dev/go-backup/internal/config"
	"github.com/0x07cf-dev/go-backup/internal/lang"
	"github.com/0x07cf-dev/go-backup/internal/logger"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ctx context.Context

var remoteDest string
var remoteRoot string

var configFile string
var envFile string
var logFile string
var logLevel string

var language string
var langFile string

var unattended bool
var simulate bool
var debug bool

var rcloneIgnoreChecksum bool
var rcloneMultiThreadSet bool
var rcloneMultiThreadStreams int
var rcloneNoTraverse bool
var rcloneNoUpdateModTime bool
var rcloneSizeOnly bool
var rcloneProgress bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-backup",
	Short: "Simple command line backup utility",
	Long:  `Go-Backup is a simple command line utility written in Go that leverages [rclone](https://rclone.org) to transfer files to cloud storage.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func remoteArg(cmd *cobra.Command, args []string) error {
	if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
		return err
	}

	// First argument is not required
	if err := cobra.MinimumNArgs(1)(cmd, args); err == nil {
		remoteDest = args[0]
	}

	// Validate remote
	if v, err := config.AsValidRemote(ctx, remoteDest, unattended); err == nil {
		remoteDest = v
	} else {
		return err
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&remoteRoot, "root", "r", "", "root backup directory on the remote")

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (defaults are .go-backup.json; configs/.go-backup.json; $HOME/.go-backup.json)")
	rootCmd.PersistentFlags().StringVarP(&envFile, "env-file", "e", "", "environment file (if none or not found, will load all variables with the prefix 'ntfy_')")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log-file", "o", "go-backup.log", "set the output logging file")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "set the minimum logging level (debug, info, warn, error, fatal)")

	rootCmd.PersistentFlags().StringVarP(&language, "lang", "l", "en", "one or more languages")
	rootCmd.PersistentFlags().StringVar(&langFile, "langFile", "", "custom language file, must end with .*.toml")

	rootCmd.PersistentFlags().BoolVarP(&unattended, "unattended", "U", false, "set this to true if you're running the program automatically. User actions will not be required")
	rootCmd.PersistentFlags().BoolVarP(&simulate, "simulate", "S", false, "simulates transfers (with fake errors)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enables debug logging")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVar(&rcloneIgnoreChecksum, "ignore-checksum", false, "")
	rootCmd.PersistentFlags().BoolVar(&rcloneMultiThreadSet, "multi-thread-set", true, "")
	rootCmd.PersistentFlags().IntVar(&rcloneMultiThreadStreams, "multi-thread-streams", 4, "")

	rootCmd.PersistentFlags().BoolVar(&rcloneNoTraverse, "no-traverse", false, "")
	rootCmd.PersistentFlags().BoolVar(&rcloneNoUpdateModTime, "no-update-modtime", false, "")

	rootCmd.PersistentFlags().BoolVar(&rcloneSizeOnly, "size-only", false, "")
	rootCmd.PersistentFlags().BoolVar(&rcloneProgress, "progress", false, "")
}

func initConfig() {
	ctx = context.Background()

	// Initialize logging
	//logLevel := logger.InfoLevel
	if debug {
		logLevel = "debug"
	}
	logger.Initialize(logFile, logLevel, unattended)

	// Find config file
	viper.SetDefault("machines", []config.Machine{})
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath("configs")
		viper.AddConfigPath(home)

		viper.SetConfigName(".go-backup")
		viper.SetConfigType("json")
	}

	// Attempt reading config file
	cfgfound := false
	if err := viper.ReadInConfig(); err == nil {
		cfgfound = true
	} else {
		// Config not found
		var e1 viper.ConfigFileNotFoundError
		var e2 viper.ConfigParseError

		if errors.As(err, &e1) || os.IsNotExist(err) {
			// Create config if not found
			logger.Infof("Setting up configuration: file %s not found. It will be created...", configFile)
			err = viper.SafeWriteConfig()
			if err != nil {
				logger.Fatalf("Error creating config file: %s", err)
			} else {
				// Read newly-created config
				if err := viper.ReadInConfig(); err == nil {
					cfgfound = true
				}
			}
		} else if errors.As(err, &e2) {
			logger.Debugf("Setting up configuration: failed to parse file %s: %s", configFile, err)
			logger.Errorf("failed to parse file %s", configFile)
		} else {
			logger.Debugf("Setting up configuration: failed to read file %s: %s", configFile, err)
			logger.Errorf("failed to read file %s", configFile)
		}
	}

	if cfgfound {
		logger.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else {
		logger.Fatal("Go-Backup cannot operate without a working configuration.")
	}

	// Environment
	envfound := false
	if envFile != "" {
		if err := loadEnvFile(envFile); err == nil {
			logger.Debugf("Using environment file: %s", envFile)
			envfound = true
		} else {
			logger.Debugf("Not using environment file: %s", err.Error())
		}
	}

	if !envfound {
		viper.SetEnvPrefix("ntfy")
		viper.AutomaticEnv()
	}

	// Language
	lang.LoadLanguages(langFile, language)

	// Log session parameters
	logger.Debugf("----------------------------------------------------------------")
	logger.Debugf("Debug: %v  |  Simulation: %v  |  Unattended: %v", debug, simulate, unattended)
	logger.Debugf("Config file: '%s'", viper.ConfigFileUsed())
	logger.Debugf("Env file: '%s'", envFile)
	logger.Debugf("Log file: '%s'", logFile)
	logger.Debugf("Language: '%s'  |  Custom lang file: %s", language, langFile)
	logger.Debugf("----------------------------------------------------------------")

	// Configure rclone
	config.ConfigureRclone(ctx, config.RcloneConfig{
		IgnoreChecksum:     rcloneIgnoreChecksum,
		MultiThreadSet:     rcloneMultiThreadSet,
		MultiThreadStreams: rcloneMultiThreadStreams,
		NoTraverse:         rcloneNoTraverse,
		NoUpdateModTime:    rcloneNoUpdateModTime,
		Progress:           rcloneProgress,
		SizeOnly:           rcloneSizeOnly,
	}, logLevel)
}

func loadEnvFile(path string) error {
	if err := godotenv.Load(path); err != nil {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		exeDir := filepath.Dir(exe)
		if err := godotenv.Load(filepath.Join(exeDir, ".env")); err != nil {
			return err
		}
	}
	return nil
}
