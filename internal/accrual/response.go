package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
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
}

type LoyaltyResponse interface {
	GetPointsByOrder(url string) (*orderResponse, error)
}

func NewLoyalty(log *zap.SugaredLogger, orderID string, accrual *float64) LoyaltyResponse {
	return &orderResponse{OrderID: orderID, log: log, Accrual: accrual}
}

var (
	ErrReadBody   = errors.New("read body")
	ErrUnmarshal  = errors.New("unmarshal response")
	ErrStatusCode = errors.New("not success status")
	ErrNotFound   = errors.New("order not found")
)

func (o *orderResponse) GetPointsByOrder(url string) (*orderResponse, error) {
	for {
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("request to loyalty: %w", err)
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
				return nil, ErrStatusCode
			}

			time.Sleep(time.Duration(delaySeconds) * time.Second)

			continue
		}

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNoContent {
				return nil, fmt.Errorf("something wents wrong: %w in url: %s", ErrNotFound, url)
			}
			o.log.Debug("сервер вернул статус-код: %d, url: %s", resp.StatusCode, url)
			return nil, ErrStatusCode
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, ErrReadBody
		}

		var orderResp orderResponse
		if err := json.Unmarshal(body, &orderResp); err != nil {
			return nil, ErrUnmarshal
		}

		return &orderResp, nil
	}
}
