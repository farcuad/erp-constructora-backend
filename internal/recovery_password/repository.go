package recoverypassword

import (
	"database/sql"
	"time"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateRecoveryToken(email, token string, expiresAt time.Time) error {
	query := `INSERT INTO password_recoveries (email, token, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, email, token, expiresAt)
	return err
}

func (r *AuthRepository) ValidateToken(email, token string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM password_recoveries 
	          WHERE email = $1 AND token = $2 AND used = FALSE AND expires_at > $3`

	err := r.db.QueryRow(query, email, token, time.Now()).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AuthRepository) MarkTokenAsUsed(email, token string) error {
	query := `UPDATE password_recoveries SET used = TRUE WHERE email = $1 AND token = $2`
	_, err := r.db.Exec(query, email, token)
	return err
}

func (r *AuthRepository) UpdateUserPassword(email, hashedPassword string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE email = $2`
	_, err := r.db.Exec(query, hashedPassword, email)
	return err
}