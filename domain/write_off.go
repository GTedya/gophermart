package domain

type WriteOff struct {
	ID         int64   `json:"-"`
	OrderID    int64   `json:"order"`
	Withdrawn  float64 `json:"sum"`
	UserID     int64   `json:"-"`
	UploadedAt string  `json:"uploaded_at"`
}
