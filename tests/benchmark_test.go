package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cevrimxe/auth-service/handlers"
	"github.com/cevrimxe/auth-service/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

func BenchmarkUserHandler_Signup(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	mockEmail := new(MockEmailService)
	handler := handlers.NewUserHandlerWithEmailService(mockRepo, mockEmail)

	// Setup mocks
	mockRepo.On("GetByEmail", mock.Anything, mock.AnythingOfType("string")).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	mockEmail.On("SendEmail", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	user := models.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}
	jsonData, _ := json.Marshal(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Signup(c)
	}
}

func BenchmarkUserHandler_Login(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	validatedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	mockRepo.On("ValidateCredentials", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(validatedUser, nil)

	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(loginData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Login(c)
	}
}

func BenchmarkUserHandler_GetMe(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	handler := handlers.NewUserHandler(mockRepo)

	user := &models.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("int64")).Return(user, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/me", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userId", int64(1))

		handler.GetMe(c)
	}
}

func BenchmarkMockUserRepository_Create(b *testing.B) {
	mockRepo := new(MockUserRepository)

	user := &models.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockRepo.Create(context.Background(), user)
	}
}

func BenchmarkMockUserRepository_GetByEmail(b *testing.B) {
	mockRepo := new(MockUserRepository)

	expectedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	mockRepo.On("GetByEmail", mock.Anything, mock.AnythingOfType("string")).Return(expectedUser, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockRepo.GetByEmail(context.Background(), "test@example.com")
	}
}

func BenchmarkMockUserRepository_ValidateCredentials(b *testing.B) {
	mockRepo := new(MockUserRepository)

	expectedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	mockRepo.On("ValidateCredentials", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(expectedUser, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockRepo.ValidateCredentials(context.Background(), "test@example.com", "password123")
	}
}
