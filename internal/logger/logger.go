package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

var (
	Logger  zerolog.Logger
	Level   zerolog.Level
	LogPath string
)

const (
	DebugLevel = zerolog.DebugLevel
	InfoLevel  = zerolog.InfoLevel
	WarnLevel  = zerolog.WarnLevel
	ErrorLevel = zerolog.ErrorLevel
	FatalLevel = zerolog.FatalLevel
)

func Initialize(path string, logLevel zerolog.Level, unattended bool) {
	// Create a file writer if a log file path is provided
	var fileWriter io.Writer
	if path != "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			fmt.Println("Error getting absolute path:", err)
		} else {
			logFile, err := os.OpenFile(abs, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("Error opening log file:", err)
			} else {
				fileWriter = logFile
				LogPath = abs
			}
		}
	}

	// Pretty logging configuration
	zerolog.TimeFieldFormat = "2006/01/02 15:04:05.000"
	zerolog.TimestampFieldName = "ts"
	zerolog.MessageFieldName = "msg"

	// If no log file path provided or encountered an error, only console will be written to
	getConsoleWriter := func() io.Writer {
		if !unattended {
			// Setup pretty logging if session is interactive
			cw := zerolog.ConsoleWriter{Out: os.Stdout}
			cw.FormatLevel = func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("| %-5s |", i))
			}
			cw.FormatMessage = func(i interface{}) string {
				return fmt.Sprintf("* %s", i)
			}
			cw.FormatFieldName = func(i interface{}) string {
				return fmt.Sprintf("%s:", i)
			}
			cw.FormatFieldValue = func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("%s", i))
			}
			return cw
		} else {
			return nil
		}
	}

	var output io.Writer
	consoleWriter := getConsoleWriter()

	if fileWriter != nil {
		if consoleWriter != nil {
			output = io.MultiWriter(consoleWriter, fileWriter)
		} else {
			output = fileWriter
		}
	} else {
		if consoleWriter != nil {
			output = consoleWriter
		} else {
			output = nil
		}
	}

	if output != nil {
		Logger = zerolog.New(output).Level(logLevel).With().Timestamp().Logger()
		Debug("Logging initialized.")
	} else {
		fmt.Println("Logging is not enabled.")
	}
}

func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debug().Msgf(format, args...)
}

func Info(msg string) {
	Logger.Info().Msg(msg)
}

func Infof(format string, args ...interface{}) {
	Logger.Info().Msgf(format, args...)
}

func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warn().Msgf(format, args...)
}

func Error(msg string) {
	Logger.Error().Msg(msg)
}

func Errorf(format string, args ...interface{}) {
	Logger.Error().Msgf(format, args...)
}

func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatal().Msgf(format, args...)
}
