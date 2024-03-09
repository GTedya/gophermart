package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/GTedya/gophermart/domain"
	"go.uber.org/zap"
)

type userRepo struct {
	DB   *sql.DB
	user *domain.User
	log  *zap.SugaredLogger
}

func NewUserRepo(DB *sql.DB, user *domain.User, log *zap.SugaredLogger) UserRepo {
	return &userRepo{DB: DB, user: user, log: log}
}

var ErrInsufficientFunds = errors.New("insufficient funds in the account")

type UserRepo interface {
	RegisterUser(ctx context.Context) (int64, error)
	GetUser(ctx context.Context) (user domain.User, err error)
	UserBalance(ctx context.Context) (current float64, withdrawn float64, err error)
	Withdraw(ctx context.Context, writeOff domain.WriteOff) error
}

func (u *userRepo) RegisterUser(ctx context.Context) (int64, error) {
	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("transaction start error: %w", err)
	}

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				u.log.Error("transaction rollback error: ", txErr)
				return
			}
		}
		if txErr := tx.Commit(); txErr != nil {
			u.log.Error("commit transaction error", txErr)
		}
	}()

	var insertedID int64

	err = tx.QueryRow("INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id",
		u.user.Login, u.user.Password).Scan(&insertedID)
	if err != nil {
		return 0, fmt.Errorf("saving url execution error: %w", err)
	}

	return insertedID, nil
}

func (u *userRepo) GetUser(ctx context.Context) (user domain.User, err error) {
	tx, err := u.DB.BeginTx(ctx, nil)

	if err != nil {
		return user, fmt.Errorf("transaction start error: %w", err)
	}

	err = tx.QueryRowContext(ctx, "SELECT id,login,password FROM users where login = $1", u.user.Login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("query error: %w", err)
	}

	return user, nil
}

func (u *userRepo) UserBalance(ctx context.Context) (current float64, withdrawn float64, err error) {
	tx, err := u.DB.BeginTx(ctx, nil)

	if err != nil {
		return 0, 0, fmt.Errorf("transaction start error: %w", err)
	}

	_ = tx.QueryRowContext(ctx, "SELECT SUM(accrual) AS total_accrual FROM order_accruals WHERE user_id =$1;", u.user.ID).Scan(&current)
	_ = tx.QueryRowContext(ctx, "SELECT SUM(withdrawn) AS total_accrual FROM write_off_history WHERE user_id =$1;", u.user.ID).Scan(&withdrawn)
	current = current - withdrawn
	return
}

func (u *userRepo) Withdraw(ctx context.Context, writeOff domain.WriteOff) error {
	balance, _, err := u.UserBalance(ctx)
	if err != nil {
		return fmt.Errorf("balance getting error: %w", err)
	}
	if balance < writeOff.Withdrawn {
		return ErrInsufficientFunds
	}
	tx, err := u.DB.BeginTx(ctx, nil)
	_, err = tx.ExecContext(ctx, "INSERT INTO write_off_history (order_id, user_id, withdrawn) values ($1, $2, $3)",
		writeOff.OrderID, writeOff.UserID, writeOff.Withdrawn)
	if err != nil {
		return fmt.Errorf("query execution error: %w", err)
	}
	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				u.log.Error("transaction rollback error: ", txErr)
				return
			}
		}
		if txErr := tx.Commit(); txErr != nil {
			u.log.Error("commit transaction error", txErr)
		}
	}()
	return nil
}
