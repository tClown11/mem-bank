package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"mem_bank/configs"
)

// Logger interface defines the logging contract
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

// LogrusLogger implements the Logger interface using logrus
type LogrusLogger struct {
	*logrus.Logger
}

// LogrusEntry wraps logrus.Entry to implement Logger interface
type LogrusEntry struct {
	*logrus.Entry
}

func NewLogger(config *configs.LoggingConfig) (Logger, error) {
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

	return &LogrusLogger{Logger: log}, nil
}

// LogrusLogger methods
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusEntry{Entry: l.Logger.WithFields(fields)}
}

func (l *LogrusLogger) WithError(err error) Logger {
	return &LogrusEntry{Entry: l.Logger.WithError(err)}
}

func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusEntry{Entry: l.Logger.WithField(key, value)}
}

// LogrusEntry methods
func (e *LogrusEntry) WithFields(fields map[string]interface{}) Logger {
	return &LogrusEntry{Entry: e.Entry.WithFields(fields)}
}

func (e *LogrusEntry) WithError(err error) Logger {
	return &LogrusEntry{Entry: e.Entry.WithError(err)}
}

func (e *LogrusEntry) WithField(key string, value interface{}) Logger {
	return &LogrusEntry{Entry: e.Entry.WithField(key, value)}
}
