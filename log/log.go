package log

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Duplicated constants from zap for more intuitive usage
const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zap.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zap.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zap.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zap.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zap.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zap.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zap.FatalLevel
)

const (
	// ProductionEnv sets production config (json encoder, etc)
	ProductionEnv = "production"
	// DevelopmentEnv adjusts config for development
	DevelopmentEnv = "development"
)

var log *zap.SugaredLogger
var logLevel *zap.AtomicLevel
var levelStr = "debug"

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

func getDefaultLoggerOrPanic() *zap.SugaredLogger {
	var err error
	if log != nil {
		return log
	}
	// default level: debug
	log, logLevel, err = NewLogger(DevelopmentEnv, levelStr, []string{"stdout"})
	if err != nil {
		panic(err)
	}
	return log
}

// NewLogger creates the logger with defined level. outputs defines the outputs where the
// logs will be sent. By default, outputs contains "stdout", which prints the
// logs at the output of the process. To add a log file as output, the path
// should be added at the outputs array. To avoid printing the logs but storing
// them on a file, can use []string{"pathtofile.log"}
func NewLogger(env, levelStr string, outputs []string) (*zap.SugaredLogger, *zap.AtomicLevel, error) {
	var level zap.AtomicLevel
	err := level.UnmarshalText([]byte(levelStr))
	if err != nil {
		return nil, nil, fmt.Errorf("error on setting log level: %s", err)
	}

	var cfg zap.Config

	if env == ProductionEnv {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.Config{
			Level:            level,
			Encoding:         "console",
			OutputPaths:      outputs,
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey: "message",

				LevelKey:    "level",
				EncodeLevel: zapcore.CapitalColorLevelEncoder,

				TimeKey:        "timestamp",
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,

				CallerKey:    "caller",
				EncodeCaller: zapcore.ShortCallerEncoder,

				// StacktraceKey: "stacktrace",
				StacktraceKey: "",
				LineEnding:    zapcore.DefaultLineEnding,
			},
		}
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, nil, err
	}
	withOptions := logger.WithOptions(zap.AddCallerSkip(1))
	return withOptions.Sugar(), &level, nil
}

// SetEnv sets logger with config depending on the env provided
func SetEnv(env string) {
	if env == ProductionEnv {
		log, logLevel, _ = NewLogger(ProductionEnv, levelStr, []string{"stdout"})
	} else {
		log, logLevel, _ = NewLogger(DevelopmentEnv, levelStr, []string{"stdout"})
	}
}

// Debug calls log.Debug
func Debug(args ...interface{}) {
	getDefaultLoggerOrPanic().Debug(args...)
}

// Info calls log.Info
func Info(args ...interface{}) {
	getDefaultLoggerOrPanic().Info(args...)
}

// Warn calls log.Warn
func Warn(args ...interface{}) {
	args = appendStackTraceMaybeArgs(args)
	getDefaultLoggerOrPanic().Warn(args...)
}

// Error calls log.Error
func Error(args ...interface{}) {
	args = appendStackTraceMaybeArgs(args)
	getDefaultLoggerOrPanic().Error(args...)
}

// Fatal calls log.Fatal
func Fatal(args ...interface{}) {
	args = appendStackTraceMaybeArgs(args)
	getDefaultLoggerOrPanic().Fatal(args...)
}

// Debugf calls log.Debugf
func Debugf(template string, args ...interface{}) {
	getDefaultLoggerOrPanic().Debugf(template, args...)
}

// Infof calls log.Infof
func Infof(template string, args ...interface{}) {
	getDefaultLoggerOrPanic().Infof(template, args...)
}

// Warnf calls log.Warnf
func Warnf(template string, args ...interface{}) {
	getDefaultLoggerOrPanic().Warnf(template, args...)
}

// Fatalf calls log.Warnf
func Fatalf(template string, args ...interface{}) {
	getDefaultLoggerOrPanic().Fatalf(template, args...)
}

// Errorf calls log.Errorf and stores the error message into the ErrorFile
func Errorf(template string, args ...interface{}) {
	getDefaultLoggerOrPanic().Errorf(template, args...)
}

// Debugw calls log.Debugw
func Debugw(template string, kv ...interface{}) {
	getDefaultLoggerOrPanic().Debugw(template, kv...)
}

// Infow calls log.Infow
func Infow(template string, kv ...interface{}) {
	getDefaultLoggerOrPanic().Infow(template, kv...)
}

// Warnw calls log.Warnw
func Warnw(template string, kv ...interface{}) {
	template = appendStackTraceMaybeKV(template, kv)
	getDefaultLoggerOrPanic().Warnw(template, kv...)
}

// Errorw calls log.Errorw
func Errorw(template string, kv ...interface{}) {
	template = appendStackTraceMaybeKV(template, kv)
	getDefaultLoggerOrPanic().Errorw(template, kv...)
}

// Fatalw calls log.Fatalw
func Fatalw(template string, kv ...interface{}) {
	template = appendStackTraceMaybeKV(template, kv)
	getDefaultLoggerOrPanic().Fatalw(template, kv...)
}

// SetLevel sets level of default logger
func SetLevel(level zapcore.Level) {
	getDefaultLoggerOrPanic() // init logger if it hasn't yet been
	logLevel.SetLevel(level)
	levelStr = level.String()
}

// SetLevelStr sets level of default logger from level name
// Valid values: debug, info, warn, error, dpanic, panic, fatal
func SetLevelStr(_levelStr string) {
	l := getDefaultLoggerOrPanic() // init logger if it hasn't yet been
	err := logLevel.UnmarshalText([]byte(_levelStr))
	if err != nil {
		l.Error("can't change log level: invalid string value provided")
		return
	}
	levelStr = _levelStr
}

// appendStackTraceMaybeArgs will append the stacktrace to the args if one of them
// is an Error
func appendStackTraceMaybeArgs(args []interface{}) []interface{} {
	for i := range args {
		if err, ok := args[i].(causer); ok {
			cause := causeWithStackTrace(err.(error))
			if stErr, ok := cause.(stackTracer); ok {
				st := stErr.StackTrace()
				for i := 0; i < len(st)-2; i++ {
					args = append(args, "\n", fmt.Sprintf("%+v", st[i]))
				}
				return args
			}
			return append(args, fmt.Sprintf("%+v", cause))
		}
	}
	return args
}

// appendStackTraceMaybeKV will append the stacktrace to the KV if one of them
// is an Error
func appendStackTraceMaybeKV(msg string, kv []interface{}) string {
	for i := range kv {
		if i%2 == 0 {
			continue
		}
		// TODO: check kv[i].(causer) and add stack trace
	}
	return msg
}

func causeWithStackTrace(err error) error {
	for err != nil {
		errCauser, ok := err.(causer)
		if !ok {
			break
		}
		cause := errCauser.Cause()
		_, ok = cause.(stackTracer)
		if !ok {
			break
		}
		err = cause
	}
	return err
}

// DebugLogger Debug* log functions
type DebugLogger interface {
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Debugw(template string, kv ...interface{})
}

// InfoLogger Info* log functions
type InfoLogger interface {
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Infow(template string, kv ...interface{})
}

// WarnLogger Warn* log functions
type WarnLogger interface {
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(template string, kv ...interface{})
}

// ErrorLogger Error* log functions
type ErrorLogger interface {
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(template string, kv ...interface{})
}

// FatalLogger Fatal* log functions
type FatalLogger interface {
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Fatalw(template string, kv ...interface{})
}

// Logger is an umbrella interface for all log functions
type Logger interface {
	DebugLogger
	InfoLogger
	WarnLogger
	ErrorLogger
	FatalLogger
}

func chiRxIDFromCtx(ctx context.Context) *zap.SugaredLogger {
	rID := middleware.GetReqID(ctx)
	l := getDefaultLoggerOrPanic()
	if rID == "" {
		return l
	}
	return l.With("request-id", rID)
}

type ctxLogger struct{ ctx context.Context }

func (c ctxLogger) Debug(args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Debug(args...)
}

func (c ctxLogger) Debugf(template string, args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Debugf(template, args...)
}

func (c ctxLogger) Debugw(template string, kv ...interface{}) {
	chiRxIDFromCtx(c.ctx).Debugw(template, kv...)
}

func (c ctxLogger) Info(args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Info(args...)
}

func (c ctxLogger) Infof(template string, args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Infof(template, args...)
}

func (c ctxLogger) Infow(template string, kv ...interface{}) {
	chiRxIDFromCtx(c.ctx).Infow(template, kv...)
}

func (c ctxLogger) Warn(args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Warn(args...)
}

func (c ctxLogger) Warnf(template string, args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Warnf(template, args...)
}

func (c ctxLogger) Warnw(template string, kv ...interface{}) {
	chiRxIDFromCtx(c.ctx).Warnw(template, kv...)
}

func (c ctxLogger) Error(args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Error(args...)
}

func (c ctxLogger) Errorf(template string, args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Errorf(template, args...)
}

func (c ctxLogger) Errorw(template string, kv ...interface{}) {
	chiRxIDFromCtx(c.ctx).Errorw(template, kv...)
}

func (c ctxLogger) Fatal(args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Fatal(args...)
}

func (c ctxLogger) Fatalf(template string, args ...interface{}) {
	chiRxIDFromCtx(c.ctx).Fatalf(template, args...)
}

func (c ctxLogger) Fatalw(template string, kv ...interface{}) {
	chiRxIDFromCtx(c.ctx).Fatalw(template, kv...)
}

// WithContext creates new logger with wrapped context that would be used
// to extract request-id value.
func WithContext(ctx context.Context) Logger {
	return ctxLogger{ctx}
}

// Sync flushes buffers to log storage.
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
