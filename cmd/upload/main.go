package main

import (
	"flag"

	"github.com/0x07cf-dev/go-backup/internal/backup"
	"github.com/0x07cf-dev/go-backup/internal/logger"
)

func main() {
	// String flags
	remoteRoot := flag.String("remoteRoot", "Backups", "Specify the root backup directory on the remote.")
	envPath := flag.String("envFile", "configs/.env", "Specify the path to the environment file")
	configPath := flag.String("configFile", "configs/config.json", "Specify the path to the config file")
	langPath := flag.String("langFile", "configs/lang.en.toml", "Specify the path to a custom language file.")
	logPath := flag.String("logFile", "latest.log", "Specify the path to the log file (will be created).")

	// Bool flags
	interactive := flag.Bool("interactive", true, "Set this to false if you're running the program automatically. User actions will not be required.")
	simulate := flag.Bool("simulate", false, "whether the backup session should be simulated")
	debug := flag.Bool("debug", false, "enables debug mode")
	flag.Parse()

	// Initialize logging
	logLevel := logger.InfoLevel
	if *debug {
		logLevel = logger.DebugLevel
	}
	logger.Initialize(*logPath, logLevel)

	args := flag.Args()
	if len(args) > 1 {
		logger.Error("Incorrect usage! To run this program:")
		logger.Error(" > upload <remote>")
		logger.Error(" > upload Drive")
		logger.Error("To see a list of options:")
		logger.Fatal(" > upload --help")
	}

	remote := ""
	if len(args) == 1 {
		remote = args[0]
	}

	session := backup.NewSession(
		backup.WithRemote(remote),
		backup.WithRemoteRoot(*remoteRoot),
		backup.WithSimulation(*simulate),
		backup.WithInteractivity(*interactive),
		backup.WithEnvFile(*envPath),
		backup.WithConfigFile(*configPath),
		backup.WithLangFile(*langPath),
	)
	session.Backup()
}
