package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"mem_bank/configs"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger(config *configs.LoggingConfig) (*Logger, error) {
	log := logrus.New()

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}
	log.SetLevel(level)

	switch config.Format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	var output io.Writer
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = file
	}
	log.SetOutput(output)

	return &Logger{Logger: log}, nil
}

func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}
