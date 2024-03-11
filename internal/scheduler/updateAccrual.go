package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/GTedya/gophermart/config"
	"github.com/GTedya/gophermart/domain"
	"github.com/GTedya/gophermart/internal/accrual"
	"github.com/GTedya/gophermart/internal/repository"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"sync"
	"time"
)

const StatusFresh = 2 * time.Second

type planner struct {
	log *zap.SugaredLogger
	db  *sql.DB
	cfg config.Config
	wg  *sync.WaitGroup
}

func NewPlanner(log *zap.SugaredLogger, db *sql.DB, cfg config.Config, wg *sync.WaitGroup) Planner {
	return planner{
		log: log,
		db:  db,
		cfg: cfg,
		wg:  wg,
	}
}

type Planner interface {
	UpdateAccrual(ctx context.Context)
}

func (p planner) UpdateAccrual(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(StatusFresh)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			orderRepo := repository.NewOrderRepo(p.db, nil, p.log)
			orders, err := orderRepo.GetOrders(ctx)
			if err != nil {
				log.Error("Status updating error", err)
			}
			p.startProcessed(ctx, orders)
		case <-ctx.Done():
			return
		}
	}
}

func (p planner) startProcessed(ctx context.Context, orders []domain.Accrual) {
	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		order := order
		go func() {
			p.orderProcessed(ctx, &order)
		}()
	}
}

func (p planner) orderProcessed(ctx context.Context, order *domain.Accrual) {
	loyalty := accrual.NewLoyalty(p.log, order.OrderID, &order.Accrual)
	orderResp, err := loyalty.GetPointsByOrder(fmt.Sprintf("%s/api/orders/%s", p.cfg.AccrualSystemAddress, order.OrderID))
	if err != nil {
		p.log.Errorw("не удалось получить данные по заказу", err)
		return
	}
	if *orderResp.Accrual == 0 {
		return
	}
	order.Accrual = *orderResp.Accrual
	orderRepo := repository.NewOrderRepo(p.db, order, p.log)

	errDB := orderRepo.UpdateAccrual(ctx)
	if errDB != nil {
		p.log.Errorw("не удалось обновить заказ", err)
		return
	}
	p.log.Debug("Статусы записей успешно обновлены")
}
