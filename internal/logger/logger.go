package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger = zap.NewNop().Sugar()

func New(level string) (*zap.SugaredLogger, error) {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	// создаём новую конфигурацию логера
	cnf := zap.NewProductionConfig()
	// устанавливаем уровень
	cnf.Level = lvl
	// устанавливаем отображение
	cnf.Encoding = "console"
	// Устанавливаем удобочитаемый формат времени
	cnf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	// создаём логер
	logger, err := cnf.Build()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
