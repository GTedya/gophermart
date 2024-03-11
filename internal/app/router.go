package app

import (
	"database/sql"
	"github.com/GTedya/gophermart/internal/handlers"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"net/http"
)

type Handler interface {
	UserRegister(c echo.Context) error
	UserLogin(c echo.Context) error
	OrderLoading(c echo.Context) error
	UserOrders(c echo.Context) error
	UserBalance(e echo.Context) error
	Withdraw(c echo.Context) error
}

type Router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

func NewRouter(log *zap.SugaredLogger, db *sql.DB, secretKey []byte) Router {
	c := echo.New()
	h := handlers.NewHandler(log, db, secretKey)

	c.Use(middleware.Logger())
	c.Use(middleware.Decompress())
	c.Use(middleware.Gzip())
	c.Use(middleware.Recover())
	initRoutes(c, h, secretKey)

	return c
}

func initRoutes(c *echo.Echo, h Handler, secretKey []byte) {
	auth := c.Group("api/user")
	auth.Use(echojwt.JWT(secretKey))
	auth.POST("/orders", h.OrderLoading)
	auth.POST("/balance/withdraw", h.Withdraw)
	auth.GET("/orders", h.UserOrders)
	auth.GET("/balance", h.UserBalance)

	c.POST("/api/user/register", h.UserRegister)
	c.POST("/api/user/login", h.UserLogin)
}
