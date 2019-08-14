package tracing

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/mitchfriedman/workflow/lib/logging"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func ConfigureCorrelatedLogger(ctx context.Context, l logging.StructuredLogger) {
	span, ok := tracer.SpanFromContext(ctx)
	if ok {
		l.WithField("dd.trace_id", span.Context().TraceID())
		l.WithField("dd.span_id", span.Context().SpanID())
	}
}

// Span is a wrapper around the ddtrace Span and an error
// with some helper methods to allow access to the underlying Span.
type Span struct {
	span          tracer.Span
	operationName string
	err           error
}

// NewServiceSpan is a helper function to create a configured Span to be
// used with Services.
func NewServiceSpan(ctx context.Context, name string) (*Span, context.Context) {
	span, newContext := tracer.StartSpanFromContext(ctx, name, tracer.SpanType("service"))
	return &Span{span: span, operationName: name}, newContext
}

// NewDBSpan is a helper function to create a configured Span to be
// used with DB repositories.
func NewDBSpan(ctx context.Context, db *gorm.DB, name string) (*Span, *gorm.DB, context.Context) {
	span, newContext := tracer.StartSpanFromContext(ctx, name, tracer.SpanType("db"))
	internalSpan := Span{span: span, operationName: name}
	db.Set(parentSpanGormKey, internalSpan)
	return &internalSpan, db, newContext
}

// RecordError sets the error on the Span to be used in Finish.
func (s *Span) RecordError(err error) {
	s.err = err
}

// Finish calls the inner span.Finish with the error, if any, on the Span.
// Note: it is a noop to provide a nil error, so the zero value of the
// error is a safe operation.
func (s *Span) Finish() {
	s.span.Finish(tracer.WithError(s.err))
}

// SetTag proxies to the span.SetTag method.
func (s *Span) SetTag(key string, value interface{}) {
	s.span.SetTag(key, value)
}
