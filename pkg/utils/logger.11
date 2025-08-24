package utils

import (
	"io"

	log "github.com/sirupsen/logrus"
)

type Logger struct {
	*log.Logger
}

func NewLogger(level log.Level, output io.Writer) *Logger {
	logger := log.New()
	logger.SetLevel(level)
	logger.SetOutput(output)
	logger.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
	})
	return &Logger{logger}
}
