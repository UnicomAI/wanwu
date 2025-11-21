package log

import (
	"errors"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log file configuration
type Config struct {
	Enable     bool    `json:"enable" mapstructure:"enable"`
	Filename   string  `json:"filename" mapstructure:"filename"` // Log file name
	Level      string  `json:"level" mapstructure:"level"`       // Log level
	LevelOp    LevelOp `json:"level_op" mapstructure:"level_op"`
	MaxSize    int     `json:"max_size" mapstructure:"max_size"`       // Log rotation size in MB
	MaxBackups int     `json:"max_backups" mapstructure:"max_backups"` // Maximum number of log backups
	MaxAge     int     `json:"max_age" mapstructure:"max_age"`         // Maximum retention days
}

type LevelOp int32

const (
	LevelLT LevelOp = -2 // less than
	LevelLE LevelOp = -1 // less equal
	LevelGE LevelOp = 0  // great equal
	LevelEQ LevelOp = 1  // equal
	LevelGT LevelOp = 2  // great than
)

// Log level precedence

var slog *zap.SugaredLogger

func init() {
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, os.Stdout, zap.DebugLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	slog = logger.Sugar()
}

// Initialize logging
func InitLog(std bool, level string, cfgs ...Config) error {
	core, err := InitLogCore(std, level, cfgs...)
	if err != nil {
		return err
	}
	slog = core
	return nil
}

func InitLogCore(std bool, level string, cfgs ...Config) (*zap.SugaredLogger, error) {
	var cores []zapcore.Core
	if std {
		var stdLevel zapcore.Level
		if err := stdLevel.UnmarshalText([]byte(level)); err != nil {
			return nil, errors.New("unsupported std log level")
		}
		core := zapcore.NewCore(getEncoder(), os.Stderr, stdLevel)
		cores = append(cores, core)
	}
	for _, cfg := range cfgs {
		if !cfg.Enable {
			continue
		}
		core, err := getZapCore(cfg)
		if err != nil {
			return nil, err
		}
		cores = append(cores, core)
	}
	logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
	return logger.Sugar(), nil
}

func getZapCore(cfg Config) (zapcore.Core, error) {
	if cfg.Filename == "" {
		return nil, errors.New("empty log filename")
	}
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, errors.New("unsupported log level")
	}
	return zapcore.NewCore(getEncoder(), getLogWriter(cfg), zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		switch cfg.LevelOp {
		case LevelLT:
			return l < level
		case LevelLE:
			return l <= level
		case LevelGE:
			return l >= level
		case LevelEQ:
			return l == level
		case LevelGT:
			return l > level
		default:
			return true
		}
	})), nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter(cfg Config) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// Log debug-level messages
func Debugf(msg string, args ...interface{}) {
	slog.Debugf(msg, args...)
}

// Log info-level messages
func Infof(msg string, args ...interface{}) {
	slog.Infof(msg, args...)
}

// Log warning-level messages
func Warnf(msg string, args ...interface{}) {
	slog.Warnf(msg, args...)
}

// Log error-level messages
func Errorf(msg string, args ...interface{}) {
	slog.Errorf(msg, args...)
}

// Log panic-level messages
func Panicf(msg string, args ...interface{}) {
	slog.Panicf(msg, args...)
}

// Log fatal-level messages
func Fatalf(msg string, args ...interface{}) {
	slog.Fatalf(msg, args...)
}

func Log() *zap.SugaredLogger {
	return slog
}
