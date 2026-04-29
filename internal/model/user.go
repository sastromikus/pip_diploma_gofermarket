package model

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID           int64
	Login        string
	PasswordHash string
}