package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

const (
	// GeektimeLogFolder ...
	GeektimeLogFolder = "geektime-downloader"
)

var (
	logger = logrus.New()
)

type customFormatter struct {
}

// Format custom logrus log format
func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Get the file and line number where the log was called
	_, filename, line, _ := runtime.Caller(7)

	// Get the script name from the full file path
	fullPathName := filepath.Base(filename)

	// Format the log message
	message := fmt.Sprintf("[%s] [%s] [%s:%d] %s\n",
		entry.Time.Format("2006-01-02 15:04:05"), // Date-time
		entry.Level.String(),                     // Log level
		fullPathName,                             // Full path name
		line,                                     // Line number
		entry.Message,                            // Log message
	)

	return []byte(message), nil
}

func init() {
	userConfigDir, _ := os.UserConfigDir()
	logDir := filepath.Join(userConfigDir, GeektimeLogFolder)
    logFilePath := filepath.Join(logDir, GeektimeLogFolder+".log")

    if err := os.MkdirAll(logDir, 0755); err != nil {
        logger.Fatalf("Failed to create log directory: %v", err)
    }

	logger.SetReportCaller(true)
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&customFormatter{})
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = logFile
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}
	logger.SetOutput(logFile)
}

// Infof wrapper logrus log.Infof
func Infof(format string, args ...interface{}) {
	logger.Logf(logrus.InfoLevel, format, args...)
}

// Warnf wrapper logrus log.Warnf
func Warnf(format string, args ...interface{}) {
	logger.Logf(logrus.WarnLevel, format, args...)
}

// Error wrapper logrus log.Error
func Error(err error, args ...interface{}) {
	if err != nil {
		logger.WithError(err).Error(args...)
	} else {
		logger.Error(args...)
	}
}
