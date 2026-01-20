package log

import (
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	// setup logrus with json formatter
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFile:  "file",
			logrus.FieldKeyFunc:  "func",
		},
	})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetReportCaller(true)

	// add trace hook
	logrus.AddHook(&TraceHook{})
}

func Init() {}

type TraceHook struct{}

func (h *TraceHook) Levels() []logrus.Level {
	// apply to all log levels
	return logrus.AllLevels
}

func (h *TraceHook) Fire(entry *logrus.Entry) error {
	// get context from entry
	ctx := entry.Context
	if ctx == nil {
		return nil
	}

	// extract span from context
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return nil
	}

	// add trace_id and span_id to log fields
	entry.Data["trace_id"] = span.SpanContext().TraceID().String()
	entry.Data["span_id"] = span.SpanContext().SpanID().String()

	return nil
}
