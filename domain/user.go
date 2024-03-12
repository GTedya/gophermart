package domain

type User struct {
	ID       int64  `json:"id"`
	Login    string `json:"login" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required,gt=8"`
}
