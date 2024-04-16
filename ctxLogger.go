package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/rs/zerolog"
	pb "github.com/xegcrbq/logger/grpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"runtime"
	"time"
)

type CtxLogger struct {
	tag     string
	span    *trace.Span
	skip    int
	traceID string
	context.Context
	extra map[string]string
}

func (l *CtxLogger) new(tag *string, fc *int) ICLg {
	ctxLogger := CtxLogger{
		span:    l.span,
		traceID: l.traceID,
		Context: l.Context,
		extra:   l.extra,
	}
	if tag != nil {
		ctxLogger.tag = *tag
	}
	if fc != nil {
		ctxLogger.skip += *fc
	}
	return &ctxLogger
}

func (l *CtxLogger) Skip(fc int) ICLg {
	return l.new(&l.tag, &fc)
}
func (l *CtxLogger) TgTag(tag string) ICLg {
	return l.new(&tag, &l.skip)
}
func (l *CtxLogger) SpanStatus(code codes.Code, msg string, args ...any) {
	(*l.span).SetStatus(code, fmt.Sprintf(msg, args...))
}
func (l *CtxLogger) SpanTag(key, value string) {
	l.extra[key] = value
	(*l.span).SetAttributes(attribute.String(key, value))
}
func (l *CtxLogger) SpanLog(msg string, args ...any) {
	(*l.span).RecordError(errors.New(fmt.Sprintf(msg, args...)))
}

func (l *CtxLogger) SpanSetKV(key string, value any) {
	bytes, err := json.Marshal(value)
	if err != nil {
		l.Error(err)
		return
	}
	(*l.span).AddEvent("", trace.WithAttributes(attribute.Key(key).String(string(bytes))))
}

func (l *CtxLogger) Tracef(msg string, args ...any) {
	zl.Trace().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.TraceLevel, msg, args...)
}
func (l *CtxLogger) Trace(msg string) {
	zl.Trace().CallerSkipFrame(l.skip).Msg(msg)
	l.send(zerolog.TraceLevel, msg)
}

// debug
func (l *CtxLogger) Debugf(msg string, args ...any) {
	zl.Debug().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.DebugLevel, msg, args...)
}
func (l *CtxLogger) Debug(msg string) {
	zl.Debug().CallerSkipFrame(l.skip).Msg(msg)
	l.send(zerolog.DebugLevel, msg)
}

// info
func (l *CtxLogger) Infof(msg string, args ...any) {
	zl.Info().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.InfoLevel, msg, args...)
}
func (l *CtxLogger) Info(msg string) {
	zl.Info().CallerSkipFrame(l.skip).Msg(msg)
	l.send(zerolog.InfoLevel, msg)
}

// warn
func (l *CtxLogger) Warnf(msg string, args ...any) {
	zl.Warn().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.WarnLevel, msg, args...)
}
func (l *CtxLogger) Warn(msg string) {
	zl.Warn().CallerSkipFrame(l.skip).Msg(msg)
	l.send(zerolog.WarnLevel, msg)
}

// error
func (l *CtxLogger) Errorf(msg string, args ...any) {
	zl.Error().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.ErrorLevel, msg, args...)
}
func (l *CtxLogger) Error(err error) {
	if err != nil {
		zl.Err(err).CallerSkipFrame(l.skip).Send()
		l.send(zerolog.ErrorLevel, "%+v", err)
	}
}
func (l *CtxLogger) ErrorD(err *error) {
	if err != nil {
		if *err != nil {
			zl.Err(*err).CallerSkipFrame(l.skip).Send()
			l.send(zerolog.ErrorLevel, "%+v", *err)
		}
	}
}

// panic
func (l *CtxLogger) Panicf(msg string, args ...any) {
	l.send(zerolog.PanicLevel, msg, args...)
	forwardSendLogs()
	zl.Panic().CallerSkipFrame(l.skip).Msgf(msg, args...)

}
func (l *CtxLogger) Panic(msg string) {
	l.send(zerolog.PanicLevel, msg)
	forwardSendLogs()
	zl.Panic().CallerSkipFrame(l.skip).Msg(msg)

}

// fatal
func (l *CtxLogger) Fatalf(msg string, args ...any) {
	l.send(zerolog.FatalLevel, msg, args...)
	forwardSendLogs()
	zl.Fatal().CallerSkipFrame(l.skip).Msgf(msg, args...)
}
func (l *CtxLogger) Fatal(msg string) {
	l.send(zerolog.TraceLevel, msg)
	forwardSendLogs()
	zl.Fatal().CallerSkipFrame(l.skip).Msg(msg)
}

func (l *CtxLogger) caller() []*pb.Caller {
	var res []*pb.Caller
	for i := 0; i < CALLERCOUNT; i++ {
		pc, filename, line, ok := runtime.Caller(l.skip + 3 + i)
		if ok {
			fun := runtime.FuncForPC(pc).Name()
			res = append(res, &pb.Caller{
				FuncName: fun,
				Source:   fmt.Sprintf("%s:%d", filename, line),
			})
		}
	}
	return res
}

func (l *CtxLogger) send(level zerolog.Level, msg string, args ...any) {
	log := &pb.Log{
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
			Nanos:   int32(time.Now().Nanosecond()),
		},
		TraceId: l.traceID,
		Msg:     fmt.Sprintf(msg, args...),
		Tag:     l.tag,
		Level:   level.String(),
		Extra:   l.extra,
	}
	if level >= zerolog.ErrorLevel {
		l.SpanLog(msg, args...)
		log.Caller = l.caller()
	}
	lg.mu.Lock()
	lg.logs.Logs = append(lg.logs.Logs, log)
	lg.mu.Unlock()
}

func (l *CtxLogger) End() {
	(*l.span).End()
}

func (l *CtxLogger) LoggerContext() context.Context {
	return l.Context
}
