package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/GTedya/gophermart/domain"
	"github.com/golang-module/carbon/v2"
	"go.uber.org/zap"
	"time"
)

type orderRepo struct {
	DB    *sql.DB
	order *domain.Accrual
	log   *zap.SugaredLogger
}

func NewOrderRepo(DB *sql.DB, order *domain.Accrual, log *zap.SugaredLogger) OrderRepo {
	return &orderRepo{DB: DB, order: order, log: log}
}

var (
	ErrExistingOrderThisUser    = errors.New("номер заказа уже был загружен этим пользователем")
	ErrExistingOrderAnotherUser = errors.New("номер заказа уже был загружен другим пользователем")
)

type OrderRepo interface {
	CreateOrder(ctx context.Context) error
	GetUserOrders(ctx context.Context) ([]domain.Accrual, error)
	GetAccrual(ctx context.Context) (order domain.Accrual, err error)
	GetOrders(ctx context.Context) ([]domain.Accrual, error)
	UpdateAccrual(ctx context.Context) error
}

func (r *orderRepo) CreateOrder(ctx context.Context) error {
	var existingUserID int64
	query := "SELECT user_id FROM order_accruals WHERE order_id = $1"
	err := r.DB.QueryRowContext(ctx, query, r.order.OrderID).Scan(&existingUserID)
	if err == nil {
		if existingUserID == r.order.UserID {
			return ErrExistingOrderThisUser
		}
		return ErrExistingOrderAnotherUser
	}

	insertQuery := "INSERT INTO order_accruals (user_id, order_id) VALUES ($1, $2)"
	_, err = r.DB.ExecContext(ctx, insertQuery, r.order.UserID, r.order.OrderID)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (r *orderRepo) GetUserOrders(ctx context.Context) ([]domain.Accrual, error) {
	var userOrders []domain.Accrual
	var accrual sql.NullFloat64
	rows, err := r.DB.QueryContext(ctx, "SELECT order_id, uploaded_at, accrual, status FROM order_accruals where user_id = $1 ORDER BY uploaded_at DESC ", r.order.UserID)
	defer func() {
		er := rows.Close()
		if er != nil {
			r.log.Error(er)
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("query executing error: %w", err)
	}
	for rows.Next() {
		var order domain.Accrual
		var uploadedAt time.Time

		if err = rows.Scan(&order.OrderID, &uploadedAt, &accrual, &order.Status); err != nil {
			return nil, fmt.Errorf("rows scan error: %w", err)
		}
		if accrual.Valid {
			order.Accrual = accrual.Float64
		}
		order.UploadedAt = carbon.Parse(uploadedAt.String()).ToRfc3339String()

		userOrders = append(userOrders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return userOrders, nil
}

func (r *orderRepo) GetAccrual(ctx context.Context) (order domain.Accrual, err error) {
	err = r.DB.QueryRowContext(ctx, "SELECT status, accrual FROM order_accruals where order_id = $1",
		r.order.OrderID).Scan(&order.Status, &order.Accrual)
	if err != nil {
		return order, fmt.Errorf("status getting error: %w", err)
	}
	return order, nil
}

func (r *orderRepo) GetOrders(ctx context.Context) ([]domain.Accrual, error) {
	var userOrders []domain.Accrual
	rows, err := r.DB.QueryContext(ctx, "SELECT id,order_id,status FROM order_accruals WHERE status !='PROCESSED' AND status != 'INVALID' ORDER BY uploaded_at ")
	defer func() {
		er := rows.Close()
		if er != nil {
			r.log.Error(er)
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("query executing error: %w", err)
	}
	for rows.Next() {
		var order domain.Accrual

		if err = rows.Scan(&order.ID, &order.OrderID, &order.Status); err != nil {
			return nil, fmt.Errorf("rows scan error: %w", err)
		}

		userOrders = append(userOrders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return userOrders, nil
}

func (r *orderRepo) UpdateAccrual(ctx context.Context) error {
	insertQuery := "UPDATE order_accruals SET accrual = $1, status = 'PROCESSED' WHERE order_id = $2"
	_, err := r.DB.ExecContext(ctx, insertQuery, r.order.Accrual, r.order.OrderID)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}
