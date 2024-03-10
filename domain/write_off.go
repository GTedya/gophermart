package domain

type WriteOff struct {
	ID         int64   `json:"-"`
	OrderID    string  `json:"order"`
	Withdrawn  float64 `json:"sum"`
	UserID     int64   `json:"-"`
	UploadedAt string  `json:"uploaded_at"`
}
