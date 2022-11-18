package main

import (
	"github.com/utahta/go-cronowriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
	"io"
	"os"
	"path/filepath"
)

type logger struct {
	zapLogger *zap.Logger
}

func newLogger(stdoutPath, stderrPath string) (*logger, error) {
	var l *zap.Logger
	var err error

	options := []zap.Option{
		zap.WithCaller(false),
		zap.AddStacktrace(zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return false
		})),
	}

	if stdoutPath == "" && stderrPath == "" { // output to console
		if l, err = zap.NewProduction(options...); err != nil {
			return nil, err
		}
		return &logger{
			zapLogger: l,
		}, nil
	}

	encoderConfig := zap.NewProductionEncoderConfig()

	if stderrPath == "" { // output all to a file
		l = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(cronowriter.MustNew(stdoutPath)),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return true
			})),
			options...,
		)
	} else { // output to 2 files
		if err = os.MkdirAll(filepath.Dir(stderrPath), os.ModePerm); err != nil {
			return nil, err
		}

		l = zap.New(
			zapcore.NewTee(
				zapcore.NewCore(
					zapcore.NewJSONEncoder(encoderConfig),
					zapcore.AddSync(cronowriter.MustNew(stdoutPath)),
					zap.LevelEnablerFunc(func(level zapcore.Level) bool {
						return level < zap.ErrorLevel
					}),
				),
				zapcore.NewCore(
					zapcore.NewJSONEncoder(encoderConfig),
					zapcore.AddSync(cronowriter.MustNew(stderrPath)),
					zap.LevelEnablerFunc(func(level zapcore.Level) bool {
						return level >= zap.ErrorLevel
					}),
				),
			),
			options...,
		)
	}

	if err = os.MkdirAll(filepath.Dir(stdoutPath), os.ModePerm); err != nil {
		return nil, err
	}

	return &logger{zapLogger: l}, nil
}

func handleFields(args []any) []zap.Field {
	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args); {
		if i == len(args)-1 {
			break
		}
		key, val := args[i], args[i+1]
		keyStr, isString := key.(string)
		if !isString {
			break
		}
		fields = append(fields, zap.Any(keyStr, val))
		i += 2
	}

	return fields
}

func (l *logger) Info(msg string, args ...any) {
	l.zapLogger.Info(msg, handleFields(args)...)
}

func (l *logger) Error(err error, msg string, args ...any) {
	l.zapLogger.Error(msg, handleFields(append(args, "error", err.Error()))...)
}

func (l *logger) stdout(kv ...any) io.Writer {
	return &zapio.Writer{
		Log:   l.zapLogger.With(handleFields(kv)...),
		Level: zapcore.InfoLevel,
	}
}

func (l *logger) stderr(kv ...any) io.Writer {
	return &zapio.Writer{
		Log:   l.zapLogger.With(handleFields(kv)...),
		Level: zapcore.ErrorLevel,
	}
}
