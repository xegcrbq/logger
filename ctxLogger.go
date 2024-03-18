package logger

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	pb "logger/grpc"
	"runtime"
	"time"
)

type CLg struct {
	tag     string
	span    *trace.Span
	skip    int
	traceID string

	context.Context
}

func (l *CLg) new(tag *string, fc *int) ICLg {
	ctxLogger := CLg{
		span:    l.span,
		traceID: l.traceID,
		Context: l.Context,
	}
	if tag != nil {
		ctxLogger.tag = *tag
	}
	if fc != nil {
		ctxLogger.skip += *fc
	}
	return &ctxLogger
}

func (l *CLg) Skip(fc int) ICLg {
	return l.new(&l.tag, &fc)
}
func (l *CLg) TgTag(tag string) ICLg {
	return l.new(&tag, &l.skip)
}
func (l *CLg) SpanStatus(code codes.Code, msg string, args ...any) {
	(*l.span).SetStatus(code, fmt.Sprintf(msg, args...))
}
func (l *CLg) SpanTag(key, value string) {
	(*l.span).SetAttributes(attribute.String(key, value))
}
func (l *CLg) SpanLog(msg string, args ...any) {
	(*l.span).RecordError(errors.New(fmt.Sprintf(msg, args...)))
}

func (l *CLg) Tracef(msg string, args ...any) {
	zl.Trace().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.TraceLevel, msg, args...)
}
func (l *CLg) Trace(msg string) {
	zl.Trace().CallerSkipFrame(l.skip).Msg(msg)
	l.send(zerolog.TraceLevel, msg)
}

// debug
func (l *CLg) Debugf(msg string, args ...any) {
	zl.Debug().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.DebugLevel, msg, args...)
}
func (l *CLg) Debug(msg string) {
	zl.Debug().CallerSkipFrame(l.skip).Msg(msg)
	l.send(zerolog.DebugLevel, msg)
}

// info
func (l *CLg) Infof(msg string, args ...any) {
	zl.Info().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.InfoLevel, msg, args...)
}
func (l *CLg) Info(msg string) {
	zl.Info().CallerSkipFrame(l.skip).Msg(msg)
	l.send(zerolog.InfoLevel, msg)
}

// warn
func (l *CLg) Warnf(msg string, args ...any) {
	zl.Warn().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.WarnLevel, msg, args...)
}
func (l *CLg) Warn(msg string) {
	zl.Warn().CallerSkipFrame(l.skip).Msg(msg)
	l.send(zerolog.WarnLevel, msg)
}

// error
func (l *CLg) Errorf(msg string, args ...any) {
	zl.Error().CallerSkipFrame(l.skip).Msgf(msg, args...)
	l.send(zerolog.ErrorLevel, msg, args...)
}
func (l *CLg) Error(err error) {
	if err != nil {
		zl.Err(err).CallerSkipFrame(l.skip).Send()
		l.send(zerolog.ErrorLevel, err.Error())
	}
}
func (l *CLg) ErrorD(err *error) {
	if err != nil {
		if *err != nil {
			zl.Err(*err).CallerSkipFrame(l.skip).Send()
			l.send(zerolog.ErrorLevel, (*err).Error())
		}
	}
}

// panic
func (l *CLg) Panicf(msg string, args ...any) {
	l.send(zerolog.PanicLevel, msg, args...)
	forwardSendLogs()
	zl.Panic().CallerSkipFrame(l.skip).Msgf(msg, args...)

}
func (l *CLg) Panic(msg string) {
	l.send(zerolog.PanicLevel, msg)
	forwardSendLogs()
	zl.Panic().CallerSkipFrame(l.skip).Msg(msg)

}

// fatal
func (l *CLg) Fatalf(msg string, args ...any) {
	l.send(zerolog.FatalLevel, msg, args...)
	forwardSendLogs()
	zl.Fatal().CallerSkipFrame(l.skip).Msgf(msg, args...)
}
func (l *CLg) Fatal(msg string) {
	l.send(zerolog.TraceLevel, msg)
	forwardSendLogs()
	zl.Fatal().CallerSkipFrame(l.skip).Msg(msg)
}

func (l *CLg) caller() []*pb.Caller {
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

func (l *CLg) send(level zerolog.Level, msg string, args ...any) {
	log := &pb.Log{
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
			Nanos:   int32(time.Now().Nanosecond()),
		},
		TraceId: l.traceID,
		Msg:     fmt.Sprintf(msg, args...),
		Tag:     l.tag,
		Caller:  l.caller(),
		Level:   level.String(),
	}
	lg.mu.Lock()
	lg.logs.Logs = append(lg.logs.Logs, log)
	lg.mu.Unlock()
}

func (l *CLg) SpanSend() {
	(*l.span).End()
}
