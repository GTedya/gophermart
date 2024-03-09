package handlers

import (
	"context"
	"database/sql"
	"errors"
	"github.com/GTedya/gophermart/domain"
	"github.com/GTedya/gophermart/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/theplant/luhn"
	"net/http"
	"time"
)

const (
	UniqueViolationErr = pq.ErrorCode("23505")
	TokenExpires       = 12 * time.Hour
)

type Token struct {
	UserID int64 `json:"user_id"`
	jwt.StandardClaims
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (h *handler) UserRegister(c echo.Context) error {
	user := new(domain.User)
	ctx := context.Background()

	if err := c.Bind(user); err != nil {
		return c.String(http.StatusInternalServerError, "bad request")
	}

	validate := validator.New()
	err := validate.Struct(user)
	if err != nil {
		return c.String(http.StatusBadRequest, "Проверьте корректность ввода данных")
	}

	user.Password, err = hashPassword(user.Password)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	repo := repository.NewUserRepo(h.db, user, h.log)

	var pqErr *pq.Error

	id, err := repo.RegisterUser(ctx)
	if ok := errors.As(err, &pqErr); ok && pqErr.Code == UniqueViolationErr {
		return c.String(http.StatusConflict, "логин уже занят")
	}

	if err != nil {
		h.log.Errorf("user registratin error: %w", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	tok, err := tokenCreate(h.secretKey, id)
	if err != nil {
		h.log.Errorf("token creating error: %w", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Writer.Header().Add("Authorization", tok)

	return c.NoContent(http.StatusOK)
}

func (h *handler) UserLogin(c echo.Context) error {
	user := new(domain.User)
	ctx := context.Background()
	if err := c.Bind(user); err != nil {
		return c.String(http.StatusInternalServerError, "bad request")
	}
	validate := validator.New()

	err := validate.Struct(user)
	if err != nil {
		return c.String(http.StatusBadRequest, "Проверьте корректность ввода данных")
	}

	repo := repository.NewUserRepo(h.db, user, h.log)
	getUser, err := repo.GetUser(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return c.String(http.StatusUnauthorized, "Проверьте корректность ввода данных")
	}
	if err != nil {
		h.log.Errorf("user getting error: %w", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if ok := checkPasswordHash(user.Password, getUser.Password); !ok {
		return c.String(http.StatusUnauthorized, "Проверьте корректность ввода данных")
	}

	tok, err := tokenCreate(h.secretKey, getUser.ID)
	if err != nil {
		h.log.Errorf("token creating error: %w", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Writer.Header().Add("Authorization", tok)

	return c.NoContent(http.StatusOK)
}

func (h *handler) UserBalance(c echo.Context) error {
	ctx := context.Background()
	var balance Balance
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
	repo := repository.NewUserRepo(h.db, &domain.User{ID: userID}, h.log)
	balance.Current, balance.Withdrawn, err = repo.UserBalance(ctx)
	if err != nil {
		h.log.Errorf("user balance getting error: %w", err)
	}

	return c.JSON(http.StatusOK, balance)
}

func (h *handler) Withdraw(c echo.Context) error {
	ctx := context.Background()
	var cancellation domain.WriteOff

	bearerToken := c.Request().Header.Get("Authorization")
	userID, err := getIDFromToken(bearerToken, h.secretKey)
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			h.log.Error("jwt signature error: ", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		h.log.Error("jwt parsing error: ", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	err = c.Bind(&cancellation)
	if err != nil {
		h.log.Errorf("binding error: %w", err)
		return c.String(http.StatusInternalServerError, "bad request")
	}
	if !luhn.Valid(int(cancellation.OrderID)) {
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	cancellation.UserID = userID

	repo := repository.NewUserRepo(h.db, &domain.User{ID: userID}, h.log)
	err = repo.Withdraw(ctx, cancellation)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientFunds) {
			return c.NoContent(http.StatusPaymentRequired)
		}
		h.log.Errorf("withdrawn error: %w", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
