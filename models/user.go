package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cevrimxe/auth-service/database"
	"github.com/cevrimxe/auth-service/utils"
	"github.com/jackc/pgx/v4"
)

type User struct {
	ID               int64      `json:"id"`
	Email            string     `json:"email" binding:"required,email"`
	Password         string     `json:"password" binding:"required"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	IsActive         bool       `json:"is_active"`
	EmailVerified    bool       `json:"email_verified"`
	Role             string     `json:"role"`
	ResetToken       *string    `json:"reset_token,omitempty"`
	ResetTokenExpiry *time.Time `json:"reset_token_expiry,omitempty"`
}

func (u *User) Save() error {
	query := `
	INSERT INTO users (
    	email,
    	password_hash,
    	first_name,
    	last_name,
    	created_at,
    	updated_at,
    	is_active,
    	email_verified,
    	role,
    	reset_token,
    	reset_token_expiry
	) VALUES (
    	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
	)
	RETURNING id
	`

	hashedPassword, err := utils.HashPassword(u.Password)

	if err != nil {
		return err
	}

	err = database.DB.QueryRow(
		context.Background(),
		query,
		u.Email,
		hashedPassword,
		u.FirstName,
		u.LastName,
		u.CreatedAt,
		u.UpdatedAt,
		u.IsActive,
		u.EmailVerified,
		u.Role,
		u.ResetToken,
		u.ResetTokenExpiry,
	).Scan(&u.ID)

	fmt.Println("userID after insert:", u.ID)

	return err
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	err := database.DB.QueryRow(context.Background(), `
		SELECT id, email, password_hash, first_name, last_name, created_at, updated_at, is_active, email_verified, role
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.EmailVerified, &user.Role,
	)
	if err != nil {
		if err == pgx.ErrNoRows { // ← doğru kontrol bu
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (u *User) ValidateCredentials() error {
	query := "SELECT id, password_hash, email_verified FROM users WHERE email = $1"
	row := database.DB.QueryRow(context.Background(), query, u.Email)

	var retrievedPassword string
	var emailVerified bool
	err := row.Scan(&u.ID, &retrievedPassword, &emailVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("no user found with this email")
		}
		return errors.New("failed to query database: " + err.Error())
	}

	passwordIsValid := utils.CheckPasswordHash(u.Password, retrievedPassword)
	if !passwordIsValid {
		return errors.New("invalid credentials")
	}

	if !emailVerified {
		return errors.New("email not verified")
	}

	return nil
}

func GetUserById(id int64) (*User, error) {
	var user User
	err := database.DB.QueryRow(context.Background(), `
		SELECT id, email, password_hash, first_name, last_name, created_at, updated_at, is_active, email_verified, role
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.EmailVerified, &user.Role,
	)
	if err != nil {
		if err == pgx.ErrNoRows { // Eğer kullanıcı bulunmazsa hata döner
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func UpdateEmailVerified(userId int64) error {
	query := `
		UPDATE users 
		SET email_verified = true 
		WHERE id = $1
	`

	_, err := database.DB.Exec(context.Background(), query, userId)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) UpdatePassword(newPassword string) error {

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	query := `
        UPDATE users
        SET password_hash = $1, updated_at = $2
        WHERE id = $3
    `

	_, err = database.DB.Exec(context.Background(), query, hashedPassword, time.Now(), u.ID)
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return nil
}
