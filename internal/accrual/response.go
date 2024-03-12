package accrual

import (
	"encoding/json"
	"fmt"
	"github.com/GTedya/gophermart/domain"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"
)

type orderResponse struct {
	log     *zap.SugaredLogger
	OrderID string   `json:"order"`
	Accrual *float64 `json:"accrual,omitempty"`
	Status  string   `json:"status"`
}

type LoyaltyResponse interface {
	GetPointsByOrder(url string, order *domain.Accrual) error
}

func NewLoyalty(log *zap.SugaredLogger, orderID string, accrual *float64) LoyaltyResponse {
	return &orderResponse{OrderID: orderID, log: log, Accrual: accrual}
}

func (o *orderResponse) GetPointsByOrder(url string, order *domain.Accrual) error {
	for {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("request to loyalty: %w", err)
		}
		defer func() {
			er := resp.Body.Close()
			if er != nil {
				o.log.Errorf("Body closing error: %w", err)
			}
		}()
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			delaySeconds, err := strconv.Atoi(retryAfter)
			if err != nil {
				o.log.Infof("ошибка при чтении заголовка Retry-After: %w", err)
				return fmt.Errorf("error during getting Retry-after header %w", err)
			}

			time.Sleep(time.Duration(delaySeconds) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNoContent {
				return fmt.Errorf("something wents wrong: %w in url: %s", err, url)
			}
			return fmt.Errorf("сервер вернул статус-код: %d, url: %s", resp.StatusCode, url)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("body reading error: %w", err)
		}

		var orderResp orderResponse
		if err = json.Unmarshal(body, &orderResp); err != nil {
			return fmt.Errorf("json unmarshalling error: %w", err)
		}
		order.Status = orderResp.Status
		if order.Accrual != 0 {
			order.Accrual = *orderResp.Accrual
		}

		return nil
	}
}
