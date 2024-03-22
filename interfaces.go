package logger

import (
	"context"
	"go.opentelemetry.io/otel/codes"
)

// context lg
type ICLg interface {
	new(tag *string, fc *int) ICLg
	Skip(fc int) ICLg
	TgTag(tag string) ICLg
	SpanStatus(code codes.Code, msg string, args ...any)
	SpanSetKV(key string, value any)
	SpanTag(key, value string)
	SpanLog(msg string, args ...any)
	End()
	LoggerContext() context.Context
	IPLg
	context.Context
}

// pure lg
type IPLg interface {
	//trace
	Tracef(msg string, args ...any)
	Trace(msg string)
	//debug
	Debugf(msg string, args ...any)
	Debug(msg string)
	//info
	Infof(msg string, args ...any)
	Info(msg string)
	//warn
	Warnf(msg string, args ...any)
	Warn(msg string)
	//error
	Errorf(msg string, args ...any)
	Error(err error)
	ErrorD(err *error)
	//panic
	Panicf(msg string, args ...any)
	Panic(msg string)
	//fatal
	Fatalf(msg string, args ...any)
	Fatal(msg string)
}
