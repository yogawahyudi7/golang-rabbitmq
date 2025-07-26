package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger represents the logging interface
type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	logger := logrus.New()

	// Set logger configuration
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
	})

	// Set default level to info
	logger.SetLevel(logrus.InfoLevel)

	// Get environment from ENV
	env := os.Getenv("GO_ENV")
	if env == "development" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	return &logrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

// Implementation of Logger interface
func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(logrus.Fields(fields)),
	}
}

func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithError(err),
	}
}
