package logging

import (
	"io"

	"github.com/sirupsen/logrus"
)

type LogEntry struct {
	*logrus.Entry
}

type StructuredLogger interface {
	logrus.FieldLogger
}

func New(application string, out io.Writer) *LogEntry {
	l := logrus.New()
	l.Out = out

	return &LogEntry{
		l.WithFields(logrus.Fields{
			"applicationName": application,
		}),
	}
}
