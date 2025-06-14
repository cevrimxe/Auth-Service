package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cevrimxe/auth-service/models"
	"github.com/cevrimxe/auth-service/repository"
	"github.com/cevrimxe/auth-service/utils"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (
		email, password_hash, first_name, last_name,
		created_at, updated_at, is_active, email_verified,
		role, reset_token, reset_token_expiry
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id`

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	return r.db.QueryRow(ctx, query,
		user.Email, hashedPassword, user.FirstName, user.LastName,
		user.CreatedAt, user.UpdatedAt, user.IsActive, user.EmailVerified,
		user.Role, user.ResetToken, user.ResetTokenExpiry,
	).Scan(&user.ID)
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password_hash, first_name, last_name, 
		       created_at, updated_at, is_active, email_verified, role
		FROM users WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.EmailVerified, &user.Role,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password_hash, first_name, last_name,
		       created_at, updated_at, is_active, email_verified, role
		FROM users WHERE email = $1`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.EmailVerified, &user.Role,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at, 
		       is_active, email_verified, role
		FROM users`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName,
			&user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.EmailVerified, &user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, updated_at = $3
		WHERE id = $4`

	_, err := r.db.Exec(ctx, query,
		user.FirstName, user.LastName, time.Now(), user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3`

	_, err := r.db.Exec(ctx, query, passwordHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return nil
}

func (r *userRepository) UpdateEmailVerified(ctx context.Context, id int64) error {
	query := `
		UPDATE users 
		SET email_verified = true, updated_at = $1
		WHERE id = $2`

	_, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update email verification: %v", err)
	}

	return nil
}

func (r *userRepository) UpdateResetToken(ctx context.Context, id int64, token string, expiry time.Time) error {
	query := `
		UPDATE users
		SET reset_token = $1, reset_token_expiry = $2, updated_at = $3
		WHERE id = $4`

	_, err := r.db.Exec(ctx, query, token, expiry, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update reset token: %v", err)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *userRepository) ValidateCredentials(ctx context.Context, email, password string) (*models.User, error) {
	query := "SELECT id, password_hash, email_verified FROM users WHERE email = $1"

	var user models.User
	var retrievedPassword string
	var emailVerified bool

	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &retrievedPassword, &emailVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no user found with this email")
		}
		return nil, errors.New("failed to query database: " + err.Error())
	}

	passwordIsValid := utils.CheckPasswordHash(password, retrievedPassword)
	if !passwordIsValid {
		return nil, errors.New("invalid credentials")
	}

	if !emailVerified {
		return nil, errors.New("email not verified")
	}

	user.Email = email
	return &user, nil
}
