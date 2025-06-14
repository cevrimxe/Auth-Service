package models

import (
	"time"
)

type User struct {
	ID               int64      `json:"id" example:"1"`                                              // Kullanıcı ID'si
	Email            string     `json:"email" binding:"required,email" example:"user@example.com"`   // Kullanıcı email adresi
	Password         string     `json:"password" binding:"required" example:"password123"`           // Kullanıcı şifresi
	FirstName        string     `json:"first_name" example:"John"`                                   // Kullanıcının adı
	LastName         string     `json:"last_name" example:"Doe"`                                     // Kullanıcının soyadı
	CreatedAt        time.Time  `json:"created_at" example:"2025-05-01T12:00:00Z"`                   // Hesap oluşturulma tarihi
	UpdatedAt        time.Time  `json:"updated_at" example:"2025-05-01T12:00:00Z"`                   // Hesap güncellenme tarihi
	IsActive         bool       `json:"is_active" example:"true"`                                    // Hesap aktif mi?
	EmailVerified    bool       `json:"email_verified" example:"false"`                              // Email doğrulandı mı?
	Role             string     `json:"role" example:"user"`                                         // Kullanıcı rolü (örneğin: user, admin)
	ResetToken       *string    `json:"reset_token,omitempty" example:"abc123"`                      // Şifre sıfırlama token'ı
	ResetTokenExpiry *time.Time `json:"reset_token_expiry,omitempty" example:"2025-05-02T12:00:00Z"` // Şifre sıfırlama token'ının son kullanma tarihi
}
