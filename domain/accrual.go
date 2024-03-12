package domain

type Accrual struct {
	ID         int64   `json:"-"`
	OrderID    string  `json:"number"`
	UserID     int64   `json:"-"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}
