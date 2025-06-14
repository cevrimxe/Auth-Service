package repository

import (
	"context"
	"time"

	"github.com/cevrimxe/auth-service/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetAll(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateEmailVerified(ctx context.Context, id int64) error
	UpdateResetToken(ctx context.Context, id int64, token string, expiry time.Time) error
	Delete(ctx context.Context, id int64) error
	ValidateCredentials(ctx context.Context, email, password string) (*models.User, error)
}
