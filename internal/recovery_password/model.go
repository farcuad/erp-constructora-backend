package recoverypassword

type PasswordRecovery struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	Used      bool      `json:"used"`
	ExpiresAt string    `json:"expires_at"`
	CreatedAt string    `json:"created_at"`
}

type RequestResetDTO struct {
	Email string `json:"email"`
}

type ResetPasswordDTO struct {
	Email       string `json:"email"`
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}