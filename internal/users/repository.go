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

	// 6. Crear suscripción trial por defecto para la nueva empresa
	querySub := `
		INSERT INTO companies_subscriptions (company_id, status, start_date, trial_end_date, price, billing_cycle, max_projects, max_users, max_storage_mb)
		VALUES ($1, 'trial', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '14 days', 0, 'monthly', 1, 3, 100)`

	_, err = tx.ExecContext(ctx, querySub, comp.ID)
	if err != nil {
		return err
	}

	// 7. Si todo salió bien, guardamos los cambios permanentemente en la DB
	return tx.Commit()
}
func (r *Repository) GetEmailUserWithDetails(ctx context.Context, email string) (*User, error) {
	query := `
        SELECT DISTINCT ON (u.id)
            u.id, 
            u.company_id, 
            u.name, 
            u.email, 
            u.password_hash, 
            u.is_active,
            COALESCE(r.name, 'Trabajador') AS role_name,
            e.id AS employee_id
        FROM users u
        LEFT JOIN user_roles ur ON u.id = ur.user_id
        LEFT JOIN roles r ON ur.role_id = r.id
        LEFT JOIN employees e ON u.id = e.user_id
        WHERE u.email = $1
        ORDER BY u.id, u.created_at DESC;
    `
	var u User
	var employeeID sql.NullString // Permite recibir UUIDs o valores NULL sin que falle Go

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.CompanyID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.IsActive,
		&u.RoleName,
		&employeeID, // Escaneo seguro para e.id
	)
	if err != nil {
		return nil, err
	}

	// Convertir el NullString al puntero *string que espera tu struct User
	if employeeID.Valid {
		u.EmployeeID = &employeeID.String
	} else {
		u.EmployeeID = nil
	}

	return &u, nil
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
