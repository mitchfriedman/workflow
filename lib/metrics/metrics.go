package metrics

import (
	"fmt"
	"net"
	"time"

	"github.com/DataDog/datadog-go/statsd"

	"github.com/mitchfriedman/workflow/lib/logging"
)

type emptySink struct {
	logger logging.StructuredLogger
}

func (e *emptySink) Write(data []byte) (n int, err error) {
	e.logger.Printf("[statsd] %q\n", string(data))
	return len(data), nil
}

func (e *emptySink) SetWriteTimeout(time.Duration) error {
	return nil
}

func (e *emptySink) Close() error {
	return nil
}

func LoadStatsd(host, port, namespace string, tags []string, logger logging.StructuredLogger) (client *statsd.Client, err error) {
	if host == "" {
		return statsd.NewWithWriter(&emptySink{logger})
	}

	return statsd.New(net.JoinHostPort(host, port),
		statsd.WithNamespace(fmt.Sprintf("%s.", namespace)),
		statsd.WithTags(tags),
	)
}
