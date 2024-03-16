package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

type LogLevel int8

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

var (
	LogPath     string
	level       LogLevel
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

// I feel like there's a better way to do this, out there in the cosmos.
// But life is short
func Initialize(path string, logLevel LogLevel) {
	var fileWriter io.Writer
	if path != "" {
		abs, err := filepath.Abs(path)
		if err == nil {
			logFile, err := os.OpenFile(abs, os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				fileWriter = logFile
				LogPath = abs
			}
		}
	}

	out := io.MultiWriter(os.Stdout, fileWriter)
	flags := log.Ldate | log.Ltime

	debugLogger = log.New(out, "[DEBUG] ", flags)
	infoLogger = log.New(out, "[INFO] ", flags)
	infoLogger.SetOutput(out)
	warnLogger = log.New(out, "[WARN] ", flags)
	errorLogger = log.New(out, "[ERROR] ", flags)
	LogPath = path
	level = logLevel
}

func logln(logger *log.Logger, l LogLevel, v ...interface{}) {
	if level <= l {
		logger.Println(v...)
	}
}

func logf(logger *log.Logger, l LogLevel, format string, v ...interface{}) {
	if level <= l {
		logger.Printf(format, v...)
	}
}

func Debug(v ...interface{}) {
	logln(debugLogger, DebugLevel, v...)
}

func Debugf(format string, v ...interface{}) {
	logf(debugLogger, DebugLevel, format, v...)
}

func Info(v ...interface{}) {
	logln(infoLogger, InfoLevel, v...)
}

func Infof(format string, v ...interface{}) {
	logf(infoLogger, InfoLevel, format, v...)
}

func Warn(v ...interface{}) {
	logln(warnLogger, WarningLevel, v...)
}

func Warnf(format string, v ...interface{}) {
	logf(warnLogger, WarningLevel, format, v...)
}

func Error(v ...interface{}) {
	logln(errorLogger, ErrorLevel, v...)
}

func Errorf(format string, v ...interface{}) {
	logf(errorLogger, ErrorLevel, format, v...)
}

func Fatal(v ...interface{}) {
	if level <= FatalLevel {
		errorLogger.SetOutput(os.Stderr)
		errorLogger.SetPrefix("[FATAL] ")
		errorLogger.Fatalln(v...)
	}
}

func Fatalf(format string, v ...interface{}) {
	if level <= FatalLevel {
		errorLogger.SetOutput(os.Stderr)
		errorLogger.SetPrefix("[FATAL] ")
		errorLogger.Fatalf(format, v...)
	}
}
