package test

import "github.com/sirupsen/logrus"

// LogWrapper wraps a logger
type LogWrapper struct {
	Logger *logrus.Logger
}

// Error logs an error.
func (l *LogWrapper) Error(msg string) {
	l.Logger.Error(msg)
}
