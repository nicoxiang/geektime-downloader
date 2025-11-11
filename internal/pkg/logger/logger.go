package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// GeektimeLogFolder ...
	GeektimeLogFolder = "geektime-downloader"
)

var logger = logrus.New()

type customFormatter struct{}

// Format custom logrus log format
func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Get the file and line number where the log was called
	_, filename, line, _ := runtime.Caller(7)

	// Get the script name from the full file path
	fullPathName := filepath.Base(filename)

	msg := entry.Message
	if errVal, ok := entry.Data["error"]; ok && errVal != nil {
		msg = fmt.Sprintf("%s | error: %v", msg, errVal)
	}

	message := fmt.Sprintf("[%s] [%s] [%s:%d] %s\n",
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Level.String(),
		fullPathName,
		line,
		msg,
	)

	return []byte(message), nil
}

func Init(level string) {
	userConfigDir, _ := os.UserConfigDir()
	logDir := filepath.Join(userConfigDir, GeektimeLogFolder)
	logFilePath := filepath.Join(logDir, GeektimeLogFolder+".log")

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		logger.Fatalf("Failed to create log directory: %v", err)
	}

	logger.SetReportCaller(true)
	logger.SetFormatter(&customFormatter{})

	switch strings.ToLower(level) {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "none":
		// discard all logs
		logger.SetOutput(io.Discard)
		return
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err == nil {
		logger.SetOutput(logFile)
	} else {
		fmt.Fprintf(os.Stderr, "Failed to log to file, using stderr\n")
		logger.SetOutput(os.Stderr)
	}
}

// Infof wrapper logrus log.Infof
func Infof(format string, args ...interface{}) {
	logger.Logf(logrus.InfoLevel, format, args...)
}

// Warnf wrapper logrus log.Warnf
func Warnf(format string, args ...interface{}) {
	logger.Logf(logrus.WarnLevel, format, args...)
}

// Errorf wrapper logrus log.Errorf
func Errorf(err error, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if err != nil {
		logger.WithError(err).Error(msg)
	} else {
		logger.Error(msg)
	}
}
