package tests

import (
	"context"
	"testing"
	"time"

	"github.com/cevrimxe/auth-service/models"
	"github.com/cevrimxe/auth-service/repository"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *pgxpool.Pool
	repo repository.UserRepository
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	// Bu testler gerçek database bağlantısı gerektirir
	// Test için ayrı bir test database kullanılmalı
	// Şimdilik mock testlerle devam ediyoruz
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *UserRepositoryTestSuite) TestCreateUser() {
	// Bu test gerçek database bağlantısı gerektirir
	// Mock testlerle devam ediyoruz
	suite.T().Skip("Integration test - requires database")
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

// Mock Repository Tests
func TestMockUserRepository_Create(t *testing.T) {
	mockRepo := new(MockUserRepository)

	user := &models.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Role:      "user",
		IsActive:  true,
	}

	mockRepo.On("Create", context.Background(), user).Return(nil)

	err := mockRepo.Create(context.Background(), user)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID) // Mock sets ID to 1
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_GetByEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)

	expectedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	mockRepo.On("GetByEmail", context.Background(), "test@example.com").Return(expectedUser, nil)

	user, err := mockRepo.GetByEmail(context.Background(), "test@example.com")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_GetByEmail_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)

	mockRepo.On("GetByEmail", context.Background(), "nonexistent@example.com").Return(nil, nil)

	user, err := mockRepo.GetByEmail(context.Background(), "nonexistent@example.com")

	assert.NoError(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_GetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)

	expectedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	mockRepo.On("GetByID", context.Background(), int64(1)).Return(expectedUser, nil)

	user, err := mockRepo.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_Update(t *testing.T) {
	mockRepo := new(MockUserRepository)

	user := &models.User{
		ID:        1,
		FirstName: "Updated",
		LastName:  "Name",
	}

	mockRepo.On("Update", context.Background(), user).Return(nil)

	err := mockRepo.Update(context.Background(), user)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_UpdatePassword(t *testing.T) {
	mockRepo := new(MockUserRepository)

	mockRepo.On("UpdatePassword", context.Background(), int64(1), "hashedpassword").Return(nil)

	err := mockRepo.UpdatePassword(context.Background(), 1, "hashedpassword")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_UpdateEmailVerified(t *testing.T) {
	mockRepo := new(MockUserRepository)

	mockRepo.On("UpdateEmailVerified", context.Background(), int64(1)).Return(nil)

	err := mockRepo.UpdateEmailVerified(context.Background(), 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_GetAll(t *testing.T) {
	mockRepo := new(MockUserRepository)

	expectedUsers := []*models.User{
		{ID: 1, Email: "user1@example.com"},
		{ID: 2, Email: "user2@example.com"},
	}

	mockRepo.On("GetAll", context.Background()).Return(expectedUsers, nil)

	users, err := mockRepo.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
	assert.Len(t, users, 2)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_ValidateCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)

	expectedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	mockRepo.On("ValidateCredentials", context.Background(), "test@example.com", "password123").Return(expectedUser, nil)

	user, err := mockRepo.ValidateCredentials(context.Background(), "test@example.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_Delete(t *testing.T) {
	mockRepo := new(MockUserRepository)

	mockRepo.On("Delete", context.Background(), int64(1)).Return(nil)

	err := mockRepo.Delete(context.Background(), 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
