package log

import (
	"github.com/sirupsen/logrus"
)

type FieldLogger interface {
	logrus.FieldLogger
	Log(level logrus.Level, args ...interface{})
}

type LogrusLogger struct {
	logger *logrus.Logger
}

var _ FieldLogger = &LogrusLogger{}

func (l *LogrusLogger) Log(level logrus.Level, args ...interface{}) {
	l.logger.Log(level, args...)
}

func (l *LogrusLogger) WithField(key string, value interface{}) *logrus.Entry {
	return l.logger.WithField(key, value)
}

func (l *LogrusLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.logger.WithFields(fields)
}

func (l *LogrusLogger) WithError(err error) *logrus.Entry {
	return l.logger.WithError(err)
}

func (l *LogrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *LogrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *LogrusLogger) Panic(args ...interface{}) {
	l.logger.Panic(args...)
}

func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *LogrusLogger) Trace(args ...interface{}) {
	l.logger.Trace(args...)
}

func (l *LogrusLogger) Print(args ...interface{}) {
	l.logger.Print(args...)
}

func (l *LogrusLogger) Tracef(format string, args ...interface{}) {
	l.logger.Tracef(format, args...)
}

func (l *LogrusLogger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *LogrusLogger) Panicf(format string, args ...interface{}) {
	l.logger.Panicf(format, args...)
}

func (l *LogrusLogger) Traceln(args ...interface{}) {
	l.logger.Traceln(args...)
}

func (l *LogrusLogger) Println(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *LogrusLogger) Debugln(args ...interface{}) {
	l.logger.Debugln(args...)
}

func (l *LogrusLogger) Infoln(args ...interface{}) {
	l.logger.Infoln(args...)
}

func (l *LogrusLogger) Warnln(args ...interface{}) {
	l.logger.Warnln(args...)
}

func (l *LogrusLogger) Errorln(args ...interface{}) {
	l.logger.Errorln(args...)
}

func (l *LogrusLogger) Fatalln(args ...interface{}) {
	l.logger.Fatalln(args...)
}

func (l *LogrusLogger) Panicln(args ...interface{}) {
	l.logger.Panicln(args...)
}

func (l *LogrusLogger) Warning(args ...interface{}) {
	l.logger.Warning(args...)
}

func (l *LogrusLogger) Warningf(format string, args ...interface{}) {
	l.logger.Warningf(format, args...)
}

func (l *LogrusLogger) Warningln(args ...interface{}) {
	l.logger.Warningln(args...)
}

func (l *LogrusLogger) SetFormatter(formatter logrus.Formatter) {
	l.logger.SetFormatter(formatter)
}

func (l *LogrusLogger) SetLevel(level logrus.Level) {
	l.logger.SetLevel(level)
}
