package sgl_logrus

import (
	"github.com/Bplotka/sgl"
	"github.com/sirupsen/logrus"
)

func toLogrusLvl(level sgl.Level) logrus.Level {
	switch level {
	case sgl.DebugLevel:
		return logrus.DebugLevel
	case sgl.InfoLevel:
		return logrus.InfoLevel
	case sgl.ErrorLevel:
		return logrus.ErrorLevel
	case sgl.FatalLevel:
		return logrus.FatalLevel
	}
	// Should not happen.
	return logrus.ErrorLevel
}

func From(s sgl.Logger) *logrus.Entry {
	l := logrus.New()
	l.Out = s.Out()
	l.Level = toLogrusLvl(s.Level())

	e := logrus.NewEntry(l)
	for _, f := range s.Fields() {
		e = e.WithField(f.Key, f.Value)
	}

	return e
}
