package superadmin

import "time"

type SuperAdmin struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
