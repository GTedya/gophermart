package logger

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"syscall"
)

var logger *zap.SugaredLogger

func init() {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	logger = log.Sugar()

	defer func(logger *zap.SugaredLogger) {
		er := logger.Sync()
		if er != nil && !errors.Is(er, syscall.ENOTTY) {
			fmt.Println(er)
		}
	}(logger)
}

func GetLogger() *zap.SugaredLogger {
	return logger
}
