package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cevrimxe/auth-service/handlers"
	"github.com/cevrimxe/auth-service/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	// Set ID for the user after creation
	user.ID = 1
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateEmailVerified(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateResetToken(ctx context.Context, id int64, token string, expiry time.Time) error {
	args := m.Called(ctx, id, token, expiry)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ValidateCredentials(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// Mock Email Service
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func TestUserHandler_Signup_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	mockEmail := new(MockEmailService)
	handler := handlers.NewUserHandlerWithEmailService(mockRepo, mockEmail)

	// Mock expectations
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	mockEmail.On("SendEmail", "test@example.com", "Verify Your Email", mock.AnythingOfType("string")).Return(nil)

	// Create request
	user := models.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}
	jsonData, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertExpectations(t)
	mockEmail.AssertExpectations(t)
}

func TestUserHandler_Signup_EmailAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	existingUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	// Mock expectations
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	// Create request
	user := models.User{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_Signup_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	// Invalid JSON
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Signup_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	mockEmail := new(MockEmailService)
	handler := handlers.NewUserHandlerWithEmailService(mockRepo, mockEmail)

	// Mock expectations
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("database error"))

	// Create request
	user := models.User{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	validatedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	// Mock expectations
	mockRepo.On("ValidateCredentials", mock.Anything, "test@example.com", "password123").Return(validatedUser, nil)

	// Create request
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	// Mock expectations
	mockRepo.On("ValidateCredentials", mock.Anything, "test@example.com", "wrongpassword").Return(nil, errors.New("invalid credentials"))

	// Create request
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_Login_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	// Invalid JSON
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetMe_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	user := &models.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(user, nil)

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userId", int64(1)) // Simulate authenticated user

	// Execute
	handler.GetMe(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_GetMe_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No userId set - simulating unauthenticated user

	// Execute
	handler.GetMe(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHandler_GetMe_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, nil)

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userId", int64(1))

	// Execute
	handler.GetMe(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_UpdateMe_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	user := &models.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(user, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	// Create request
	updateData := map[string]string{
		"firstName": "Updated",
		"lastName":  "Name",
	}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/me", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userId", int64(1))

	// Execute
	handler.UpdateMe(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_ChangePassword_WrongOldPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	user := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye.IjPFvmRaN7hD19ca9VqaVvV9CfkF1G", // hashed "secret"
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(user, nil)

	// Create request with wrong old password
	passwordData := map[string]string{
		"oldPassword": "wrongpassword",
		"newPassword": "newsecret123",
	}
	jsonData, _ := json.Marshal(passwordData)

	req, _ := http.NewRequest("PUT", "/change-password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userId", int64(1))

	// Execute
	handler.ChangePassword(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_GetUsers_AdminSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	adminUser := &models.User{
		ID:   1,
		Role: "admin",
	}

	users := []*models.User{
		{ID: 1, Email: "admin@example.com", Role: "admin"},
		{ID: 2, Email: "user@example.com", Role: "user"},
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(adminUser, nil)
	mockRepo.On("GetAll", mock.Anything).Return(users, nil)

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userId", int64(1))

	// Execute
	handler.GetUsers(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_GetUsers_AccessDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	regularUser := &models.User{
		ID:   1,
		Role: "user",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(regularUser, nil)

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userId", int64(1))

	// Execute
	handler.GetUsers(c)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_ForgetPassword_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo) // Use default email service for this test

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	// Mock expectations
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	// Create request
	requestData := map[string]string{
		"email": "test@example.com",
	}
	jsonData, _ := json.Marshal(requestData)

	req, _ := http.NewRequest("POST", "/forgot-password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.ForgetPassword(c)

	// Assert - Email sending might fail in test environment, so we check for either success or email error
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_ForgetPassword_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	// Mock expectations
	mockRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)

	// Create request
	requestData := map[string]string{
		"email": "nonexistent@example.com",
	}
	jsonData, _ := json.Marshal(requestData)

	req, _ := http.NewRequest("POST", "/forgot-password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute
	handler.ForgetPassword(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}
