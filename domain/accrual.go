package domain

type Accrual struct {
	ID         int64   `json:"-"`
	OrderID    int64   `json:"order_id"`
	UserID     int64   `json:"-"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}
