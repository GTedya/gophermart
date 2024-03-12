package middlewares

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"net/http"
	"strconv"
	"time"
)

var (
	ipRateLimiter *limiter.Limiter
	store         limiter.Store
)

const Period = 60 * time.Second
const QueryLimit = 5

func (m *middleware) IPRateLimit() echo.MiddlewareFunc {
	rate := limiter.Rate{
		Period: Period,
		Limit:  QueryLimit,
	}
	store = memory.NewStore()
	ipRateLimiter = limiter.New(store, rate)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()
			limiterCtx, err := ipRateLimiter.Get(c.Request().Context(), ip)
			if err != nil {
				m.log.Errorf("IPRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, c.Request().URL)
				return c.NoContent(http.StatusInternalServerError)
			}

			h := c.Response().Header()
			h.Set("Retry-After", strconv.Itoa(60))

			if limiterCtx.Reached {
				m.log.Info("Too Many Requests from %s on %s", ip, c.Request().URL)
				return c.JSON(http.StatusTooManyRequests, fmt.Sprint("No more than ", QueryLimit, " requests per minute allowed"))
			}

			return next(c)
		}
	}
}
