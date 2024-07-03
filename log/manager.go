package log

import (
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mpapenbr/iracelog-cli/config"
)

type LoggerManager struct {
	baseLogger      *Logger
	config          *Config
	loggers         map[string]*Logger
	mu              sync.Mutex
	defaultLogLevel zapcore.Level
}

var loggerManager *LoggerManager

func InitLoggerManager(cliArgs *config.CliArgs) (*LoggerManager, error) {
	cfg, err := LoadConfig(cliArgs.LogConfig)
	if err != nil {
		return nil, err
	}
	if cliArgs.LogLevel != "" {
		cfg.DefaultLevel = cliArgs.LogLevel
	}
	defaultLvl, err := ParseLevel(cfg.DefaultLevel)
	if err != nil {
		return nil, err
	}
	loggerManager, err = NewLoggerManager(cfg, defaultLvl)
	if err != nil {
		return nil, err
	}

	return loggerManager, nil
}

func GetLoggerManager() *LoggerManager {
	return loggerManager
}

//nolint:lll // readibility
func NewLoggerManager(cfg *Config, defaultLogLevel zapcore.Level) (*LoggerManager, error) {
	// Build the base logger

	baseLogger, err := createLogger("", cfg.DefaultLevel, defaultLogLevel)
	if err != nil {
		return nil, err
	}

	return &LoggerManager{
		baseLogger:      &Logger{l: baseLogger, level: baseLogger.Level()},
		config:          cfg,
		loggers:         make(map[string]*Logger),
		mu:              sync.Mutex{},
		defaultLogLevel: defaultLogLevel,
	}, nil
}

//nolint:lll // readibility
func createLogger(name, level string, defaultLogLevel zapcore.Level) (*zap.Logger, error) {
	baseCfg := zap.NewProductionConfig()
	baseCfg.EncoderConfig.TimeKey = "timestamp"
	baseCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Set default log level
	defaultLevel := GetZapLevel(level, defaultLogLevel)
	baseCfg.Level = zap.NewAtomicLevelAt(defaultLevel)

	// Build the base logger
	baseLogger, err := baseCfg.Build(AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	if name != "" {
		return baseLogger.Named(name), nil
	} else {
		return baseLogger, nil
	}
}

func (lm *LoggerManager) GetDefaultLogger() *Logger {
	return lm.baseLogger
}

func (lm *LoggerManager) GetLogger(name string) *Logger {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if logger, exists := lm.loggers[name]; exists {
		return logger
	}

	// Create a new logger with the default level

	newLogger, _ := createLogger(name, lm.config.Loggers[name], lm.defaultLogLevel)
	ret := &Logger{l: newLogger, level: newLogger.Level()}

	lm.loggers[name] = ret
	return ret
}

func (lm *LoggerManager) GetRegisteredLoggers() []string {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	names := make([]string, 0, len(lm.loggers))
	for name := range lm.loggers {
		names = append(names, name)
	}
	return names
}

func GetZapLevel(level string, defaultLogLevel zapcore.Level) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return defaultLogLevel
	}
}
