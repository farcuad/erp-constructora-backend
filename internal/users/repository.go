package users

import (
	"context"
	"database/sql"

	"errors"

	"golang.org/x/crypto/bcrypt"
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
        SELECT 
            u.id, 
            u.company_id, 
            u.name, 
            u.email, 
            u.password_hash, 
            u.is_active,
            COALESCE(r.name, 'Sin Rol') AS role_name
        FROM users u
        LEFT JOIN user_roles ur ON u.id = ur.user_id
        LEFT JOIN roles r ON ur.role_id = r.id
        WHERE u.email = $1
        LIMIT 1;
    `

	var u User

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.CompanyID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.IsActive,
		&u.RoleName,
	)
	if err != nil {
		return nil, err
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

func (r *Repository) GetRolesByCompanyID(ctx context.Context, companyID string) ([]Role, error) {
	query := `
		SELECT id, company_id, name, COALESCE(description, '')
		FROM roles
		WHERE company_id = $1 AND name != 'Administrador'
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.CompanyID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// Listar todos los usuarios de la empresa EXCLUYENDO el rol 'Administrador'
func (r *Repository) GetUsersExcludingAdmin(ctx context.Context, companyID string) ([]UserResponse, error) {
	query := `
		SELECT 
			u.id, 
			u.name, 
			u.email, 
			COALESCE(r.name, 'Sin Rol') AS role_name
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE u.company_id = $1 AND r.name != 'Administrador'
		ORDER BY u.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserResponse
	for rows.Next() {
		var u UserResponse
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

// Crear un nuevo Usuario asignándole su Rol en una transacción
func (r *Repository) CreateUserTransaction(ctx context.Context, companyID string, dto CreateUserDTO, hashedPassword string) (*UserResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. INSERT en 'users'
	queryUser := `
		INSERT INTO users (company_id, name, email, password_hash, is_active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id`

	var userID string
	err = tx.QueryRowContext(ctx, queryUser, companyID, dto.Name, dto.Email, hashedPassword).Scan(&userID)
	if err != nil {
		return nil, err
	}

	// 2. INSERT en 'user_roles'
	queryUserRole := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, queryUserRole, userID, dto.RoleID)
	if err != nil {
		return nil, err
	}

	// 3. Obtener el nombre del rol asignado para la respuesta
	var roleName string
	err = tx.QueryRowContext(ctx, `SELECT name FROM roles WHERE id = $1`, dto.RoleID).Scan(&roleName)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:    userID,
		Name:  dto.Name,
		Email: dto.Email,
		Role:  roleName,
	}, nil
}
func (r *Repository) UpdateUserTransaction(ctx context.Context, companyID string, dto UpdateUserDTO) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var result sql.Result

	// 1. UPDATE en 'users'
	if dto.Password != nil && *dto.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*dto.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		query := `
            UPDATE users 
            SET name = $1, email = $2, password_hash = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP 
            WHERE id = $5 AND company_id = $6`
		result, err = tx.ExecContext(ctx, query, dto.Name, dto.Email, string(hashed), dto.IsActive, dto.ID, companyID)
	} else {
		query := `
            UPDATE users 
            SET name = $1, email = $2, is_active = $3, updated_at = CURRENT_TIMESTAMP 
            WHERE id = $4 AND company_id = $5`
		result, err = tx.ExecContext(ctx, query, dto.Name, dto.Email, dto.IsActive, dto.ID, companyID)
	}

	if err != nil {
		return err
	}

	// Validar si el usuario existía y pertenecía a la empresa
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("el usuario no existe o no pertenece a esta empresa")
	}

	// 2. Reasignar Rol en 'user_roles'
	_, err = tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, dto.ID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, dto.ID, dto.RoleID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Eliminar un usuario de la empresa
func (r *Repository) DeleteUser(ctx context.Context, companyID, userID string) error {
	query := `DELETE FROM users WHERE id = $1 AND company_id = $2`
	result, err := r.db.ExecContext(ctx, query, userID, companyID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("el usuario no existe o no pertenece a esta empresa")
	}

	return nil
}
