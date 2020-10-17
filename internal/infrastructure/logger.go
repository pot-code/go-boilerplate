package infra

import (
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig options used in creating zap logger
type LoggingConfig struct {
	FilePath string // log file path
	Level    string // global logging level
	Env      string // app environment
	AppID    string
}

// Logger global logger object
// var Logger *zap.Logger

// NewLogger returns a zap logger instance based on given options.
// It's hard to extract a common interface for structured logger like zap,
// since each argument of the log function should be zap.Field type,
// it won't be nice to implement another zap
func NewLogger(cfg *LoggingConfig) (*zap.Logger, error) {
	var (
		core zapcore.Core
		err  error
	)
	switch cfg.Env {
	case "development":
		core, err = createDevLogger(cfg)
	case "production":
		core, err = createProductionLogger(cfg)
	default:
		core, err = createDevLogger(cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to create logger core: %w", err)
	}

	logger := zap.New(core, zap.AddStacktrace(zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
		return lv > zap.WarnLevel
	})), zap.AddCaller())
	return logger, nil
}

func getZapLoggingLevel(level string) (lv zapcore.Level) {
	switch level {
	case "debug":
		lv = zap.DebugLevel
	case "info":
		lv = zap.InfoLevel
	case "warn":
		lv = zap.WarnLevel
	case "error":
		lv = zap.ErrorLevel
	default:
		log.Fatal(fmt.Errorf("Unknown logging level: %s", level))
	}
	return
}

func createDevLogger(cfg *LoggingConfig) (zapcore.Core, error) {
	logEnabler := getLevelEnabler(cfg)
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.CallerKey = "log.origin.file.name"
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	if cfg.FilePath != "" {
		output, err := getFileSyncer(cfg)
		return zapcore.NewCore(encoder, output, logEnabler), err
	}
	return zapcore.NewCore(encoder, os.Stderr, logEnabler), nil
}

func createProductionLogger(cfg *LoggingConfig) (zapcore.Core, error) {
	logEnabler := getLevelEnabler(cfg)
	elkEncoderConfig := zap.NewProductionEncoderConfig()
	elkEncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z"))
	})
	elkEncoderConfig.TimeKey = "@timestamp"
	elkEncoderConfig.MessageKey = "message"
	elkEncoderConfig.LevelKey = "log.level"
	elkEncoderConfig.CallerKey = "log.origin.file.name"
	elkEncoderConfig.StacktraceKey = "error.stack_trace"
	elkEncoder := zapcore.NewJSONEncoder(elkEncoderConfig)

	if cfg.FilePath != "" {
		elkOutput, err := getFileSyncer(cfg)
		return zapcore.NewCore(elkEncoder, elkOutput, logEnabler), err
	}
	return zapcore.NewCore(elkEncoder, os.Stderr, logEnabler), nil
}

func getFileSyncer(cfg *LoggingConfig) (zapcore.WriteSyncer, error) {
	fd, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	return fd, err
}

func getLevelEnabler(cfg *LoggingConfig) zapcore.LevelEnabler {
	level := getZapLoggingLevel(cfg.Level)
	return zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
		return lv >= level
	})
}
