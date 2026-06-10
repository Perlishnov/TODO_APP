package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger(level logrus.Level) *logrus.Logger  {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00" ,
	})
	return logger
}