package logger

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	// GeektimeLogFolder ...
	GeektimeLogFolder = "geektime-downloader"
)

var (
	logger = logrus.New()
)

func init(){
	userConfigDir, _ := os.UserConfigDir()
	logFilePath := filepath.Join(userConfigDir, GeektimeLogFolder, GeektimeLogFolder + ".log")

	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetReportCaller(false)
	logger.SetLevel(logrus.WarnLevel)
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = logFile
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}
	logger.SetOutput(logFile)
}

// Trace wrapper logrus log.Trace
func Trace(args ...interface{}) {
	logger.Log(logrus.TraceLevel, args...)
}

// Debug wrapper logrus log.Debug
func Debug(args ...interface{}) {
	logger.Log(logrus.DebugLevel, args...)
}

// Info wrapper logrus log.Info
func Info(args ...interface{}) {
	logger.Log(logrus.InfoLevel, args...)
}

// Warn wrapper logrus log.Warn
func Warn(args ...interface{}) {
	logger.Log(logrus.WarnLevel, args...)
}

// Error wrapper logrus log.Error
func Error(err error, args ...interface{}){
	if err != nil{
		logger.WithError(err).Error(args...)
	}else{
		logger.Error(args...)
	}
}