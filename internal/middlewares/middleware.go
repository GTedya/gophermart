package middlewares

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type middleware struct {
	log *zap.SugaredLogger
}

type Middleware interface {
	IPRateLimit() echo.MiddlewareFunc
}

func NewMiddleware(log *zap.SugaredLogger) Middleware {
	return &middleware{log: log}
}
