package users

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ExecRegistryTransaction ejecuta los múltiples INSERTs de forma atómica
func (r *Repository) ExecRegistryTransaction(ctx context.Context, comp *Company, admin *User, defaultRoles []string) error {
	// 1. Iniciar la transacción SQL
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Defer ejecutará un Rollback si la función termina con error sin haber hecho Commit
	defer tx.Rollback()

	// 2. INSERT en 'companies'
	queryCompany := `
		INSERT INTO companies (name, nit) 
		VALUES ($1, $2) 
		RETURNING id, created_at`
	err = tx.QueryRowContext(ctx, queryCompany, comp.Name, comp.NIT).Scan(&comp.ID, &comp.CreatedAt)
	if err != nil {
		return err
	}

	// Asignamos el ID de la empresa recién creada al usuario administrador
	admin.CompanyID = comp.ID

	// 3. INSERT de Roles por defecto de la constructora (Administrador, Ingeniero, etc.)
	queryRole := `INSERT INTO roles (company_id, name) VALUES ($1, $2) RETURNING id`
	var adminRoleID string

	for _, roleName := range defaultRoles {
		var roleID string
		err := tx.QueryRowContext(ctx, queryRole, comp.ID, roleName).Scan(&roleID)
		if err != nil {
			return err
		}
		// Guardamos el ID del rol Administrador para asignárselo al usuario luego
		if roleName == "Administrador" {
			adminRoleID = roleID
		}
	}

	// 4. INSERT en 'users' (El administrador de la empresa)
	queryUser := `
		INSERT INTO users (company_id, name, email, password_hash, is_active) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created_at`
	err = tx.QueryRowContext(ctx, queryUser, admin.CompanyID, admin.Name, admin.Email, admin.PasswordHash, admin.IsActive).
		Scan(&admin.ID, &admin.CreatedAt)
	if err != nil {
		return err
	}

	// 5. INSERT en 'user_roles' (Relación muchos a muchos)
	queryUserRole := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, queryUserRole, admin.ID, adminRoleID)
	if err != nil {
		return err
	}

	// 6. Si todo salió bien, guardamos los cambios permanentemente en la DB
	return tx.Commit()
}

// Verificar si el email ya existe globalmente
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *Repository) GetEmailUser(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, company_id, name, email, password_hash, is_active, created_at FROM users WHERE email = $1`
	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.CompanyID, &user.Name, &user.Email, &user.PasswordHash, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
