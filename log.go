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
	zapLogger   *zap.Logger
	sugarLogger *zap.SugaredLogger
}

func newLogger(stdoutPath, stderrPath string) (*logger, error) {
	var l *zap.Logger
	var err error

	if stdoutPath == "" && stderrPath == "" { // output to console
		if l, err = zap.NewDevelopment(); err != nil {
			return nil, err
		}
		return &logger{
			zapLogger:   l,
			sugarLogger: l.Sugar(),
		}, nil
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if stderrPath == "" { // output all to a file
		l = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(cronowriter.MustNew(stdoutPath)),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return true
			}),
		))
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
		)
	}

	if err = os.MkdirAll(filepath.Dir(stdoutPath), os.ModePerm); err != nil {
		return nil, err
	}

	return &logger{zapLogger: l, sugarLogger: l.Sugar()}, nil
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
	l.zapLogger.Info(msg, handleFields(append(args, "error", err.Error()))...)
}

func (l *logger) Infof(format string, args ...any) {
	l.sugarLogger.Infof(format, args...)
}

func (l *logger) Errorf(format string, args ...any) {
	l.sugarLogger.Errorf(format, args...)
}

func (l *logger) stdout() io.Writer {
	return &zapio.Writer{
		Log:   l.zapLogger,
		Level: zapcore.InfoLevel,
	}
}

func (l *logger) stderr() io.Writer {
	return &zapio.Writer{
		Log:   l.zapLogger,
		Level: zapcore.ErrorLevel,
	}
}
