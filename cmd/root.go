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
	"os"
	"path/filepath"

	"github.com/0x07cf-dev/go-backup/internal/config"
	"github.com/0x07cf-dev/go-backup/internal/lang"
	"github.com/0x07cf-dev/go-backup/internal/logger"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string
var envFile string

var languages []string
var langFile string

var logFile string

var interactive bool
var simulate bool
var debug bool

var root string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-backup",
	Short: "A brief description of your application",
	Long: `Go-Backup is a simple utility written in Go that uses rclone to transfer files to a remote destination.

It follows a .json configuration in which you can define custom behaviour for each device you run it on.
You can specify which directories and/or files to transfer, along with pre and/or post-transfer commands to be executed on the machine.

It can optionally be configured to send status notifications to the user via [ntfy.sh](https://ntfy.sh/app), and/or heartbeat signals to external uptime monitoring services in order to keep track of non-interactive executions.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (defaults are .go-backup.json; .configs/.go-backup.json; $HOME/.go-backup.json)")
	rootCmd.PersistentFlags().StringVarP(&envFile, "envFile", "e", "configs/.env", "environment file")

	rootCmd.PersistentFlags().StringArrayVarP(&languages, "lang", "l", []string{"en"}, "one or more languages")
	rootCmd.PersistentFlags().StringVar(&langFile, "langFile", "", "custom language file, must end with .*.toml")

	rootCmd.PersistentFlags().StringVarP(&logFile, "logFile", "o", "go-backup.log", "output log file")
	rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", true, "set this to false if you're running the program automatically. User actions will not be required")
	rootCmd.PersistentFlags().BoolVarP(&simulate, "simulate", "s", false, "simulates transfers (with fake errors)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enables debug mode")

	rootCmd.PersistentFlags().StringVarP(&root, "root", "r", "Backups", "root backup directory on the remote")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	// Initialize logging
	logLevel := logger.InfoLevel
	if debug {
		logLevel = logger.DebugLevel
	}
	logger.Initialize(logFile, logLevel)

	// Find config file
	viper.SetDefault("machines", []config.Machine{})
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath("configs")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)

		viper.SetConfigName(".go-backup")
		viper.SetConfigType("json")
	}

	// Attempt reading config file.
	if err := viper.ReadInConfig(); err == nil {
		logger.Info("Using config file:", viper.ConfigFileUsed())
	} else {
		// Could not load
		logger.Error(err)
		switch e := err.(type) {
		case *os.PathError, viper.ConfigFileNotFoundError:
			err = viper.SafeWriteConfig()
			if err != nil {
				logger.Fatalf("Error creating config file: %s", err)
			} else {
				// Read newly-created config
				if err := viper.ReadInConfig(); err == nil {
					logger.Info("Using config file:", viper.ConfigFileUsed())
				} else {
					logger.Fatal("No config.")
				}
			}
		default:
			logger.Errorf("Unhandled: %s", e)
		}
	}

	// Environment
	envfound := false
	if envFile != "" {
		if err := loadEnvFile(envFile); err == nil {
			logger.Info("Using environment file:", envFile)
			envfound = true
		} else {
			logger.Errorf("There was an error loading environment file: %v", err)
		}
	}

	if !envfound {
		viper.SetEnvPrefix("ntfy")
		viper.AutomaticEnv()
	}

	// Language
	lang.LoadLanguages(langFile, languages...)
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
