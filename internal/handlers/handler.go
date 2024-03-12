package handlers

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler interface {
	UserRegister(c echo.Context) error
	UserLogin(c echo.Context) error
	OrderLoading(c echo.Context) error
	UserOrders(c echo.Context) error
	UserBalance(e echo.Context) error
	Withdraw(c echo.Context) error
	WithdrawHistory(c echo.Context) error
}

type handler struct {
	log       *zap.SugaredLogger
	db        *sql.DB
	secretKey []byte
}

func NewHandler(log *zap.SugaredLogger, db *sql.DB, secretKey []byte) Handler {
	return &handler{log: log, db: db, secretKey: secretKey}
}
