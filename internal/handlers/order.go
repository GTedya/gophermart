package handlers

import (
	"context"
	"database/sql"
	"errors"
	"github.com/GTedya/gophermart/domain"
	"github.com/GTedya/gophermart/internal/repository"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/theplant/luhn"
	"net/http"
	"strconv"
)

type pointCalculation struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (h *handler) OrderLoading(c echo.Context) error {
	var order domain.Accrual
	ctx := context.Background()

	bearerToken := c.Request().Header.Get("Authorization")
	userID, err := getIDFromToken(bearerToken, h.secretKey)
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			h.log.Error("jwt signature error: ", err)
			return c.NoContent(500)
		}
		return c.NoContent(500)
	}
	err = c.Bind(&order)
	if err != nil {
		h.log.Error("binding error: ", err)
		return c.NoContent(400)
	}
	if !luhn.Valid(int(order.ID)) {
		return c.NoContent(http.StatusUnprocessableEntity)
	}
	order.UserID = userID

	orderRepo := repository.NewOrderRepo(h.db, &order, h.log)

	err = orderRepo.CreateOrder(ctx)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrExistingOrderAnotherUser):
			return c.JSON(http.StatusConflict, "Номер заказа уже был загружен")
		case errors.Is(err, repository.ErrExistingOrderThisUser):
			return c.NoContent(http.StatusOK)
		default:
			h.log.Errorf("order loading error: %w", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *handler) UserOrders(c echo.Context) error {
	ctx := context.Background()
	bearerToken := c.Request().Header.Get("Authorization")
	userID, err := getIDFromToken(bearerToken, h.secretKey)
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			h.log.Error("jwt signature error: ", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		h.log.Error("jwt parsing error: ", err)
		return c.NoContent(500)
	}

	orderRepo := repository.NewOrderRepo(h.db, &domain.Accrual{UserID: userID}, h.log)
	userOrders, err := orderRepo.GetUserOrders(ctx)
	if err != nil {
		h.log.Error("order getting errors: ", err)
		return c.NoContent(500)
	}
	if len(userOrders) == 0 {
		return c.NoContent(http.StatusNoContent)

	}

	return c.JSON(http.StatusOK, userOrders)
}

func (h *handler) PointsCalculation(c echo.Context) error {
	var calculation pointCalculation
	ctx := context.Background()

	calculation.Order = c.Param("number")
	orderID, err := strconv.ParseInt(calculation.Order, 10, 64)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	orderRepo := repository.NewOrderRepo(h.db, &domain.Accrual{OrderID: orderID}, h.log)
	accrual, err := orderRepo.GetAccrual(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return c.String(http.StatusUnauthorized, "Проверьте корректность ввода данных")
	}
	if err != nil {
		h.log.Errorf("user getting error: %w", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	calculation.Accrual = accrual.Accrual
	calculation.Status = accrual.Status

	return c.JSON(http.StatusOK, calculation)
}
