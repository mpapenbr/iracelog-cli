package log

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mpapenbr/iracelog-cli/config"
)

type LoggerManager struct {
	baseLogger *Logger
	config     *Config
	loggers    map[string]*Logger
	mu         sync.Mutex
}

var loggerManager *LoggerManager

func InitLoggerManager(cliArgs *config.CliArgs) (*LoggerManager, error) {
	cfg, err := LoadConfig(cliArgs.LogConfig)
	if err != nil {
		return nil, err
	}

	loggerManager, err = NewLoggerManager(cfg)
	if err != nil {
		return nil, err
	}

	return loggerManager, nil
}

func GetLoggerManager() *LoggerManager {
	return loggerManager
}

func NewLoggerManager(cfg *Config) (*LoggerManager, error) {
	// Build the base logger
	baseLogger, err := createLogger("", cfg.DefaultLevel)
	if err != nil {
		return nil, err
	}

	return &LoggerManager{
		baseLogger: &Logger{l: baseLogger, level: baseLogger.Level()},
		config:     cfg,
		loggers:    make(map[string]*Logger),
		mu:         sync.Mutex{},
	}, nil
}

func createLogger(name, level string) (*zap.Logger, error) {
	baseCfg := zap.NewProductionConfig()
	baseCfg.EncoderConfig.TimeKey = "timestamp"
	baseCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Set default log level
	defaultLevel := GetZapLevel(level)
	baseCfg.Level = zap.NewAtomicLevelAt(defaultLevel)

	// Build the base logger
	baseLogger, err := baseCfg.Build()
	if err != nil {
		return nil, err
	}
	if name != "" {
		return baseLogger.Named(name), nil
	} else {
		return baseLogger, nil
	}
}

func (lm *LoggerManager) GetLogger(name string) *Logger {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if logger, exists := lm.loggers[name]; exists {
		return logger
	}

	// Create a new logger with the default level

	newLogger, _ := createLogger(name, lm.config.Loggers[name])
	ret := &Logger{l: newLogger, level: GetZapLevel(lm.config.Loggers[name])}

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
