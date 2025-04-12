package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func initLogger() {
	log = logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func GetLogger() *logrus.Logger {
	if log == nil {
		initLogger()
	}
	return log
}
