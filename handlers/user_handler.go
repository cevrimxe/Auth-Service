package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cevrimxe/auth-service/models"
	"github.com/cevrimxe/auth-service/repository"
	"github.com/cevrimxe/auth-service/utils"
	"github.com/gin-gonic/gin"
)

type EmailService interface {
	SendEmail(to, subject, body string) error
}

type DefaultEmailService struct{}

func (e *DefaultEmailService) SendEmail(to, subject, body string) error {
	mailer := models.NewMailer()
	return mailer.Mailer(to, subject, body)
}

type UserHandler struct {
	userRepo     repository.UserRepository
	emailService EmailService
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo:     userRepo,
		emailService: &DefaultEmailService{},
	}
}

func NewUserHandlerWithEmailService(userRepo repository.UserRepository, emailService EmailService) *UserHandler {
	return &UserHandler{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

// @Summary Sign up a new user
// @Description Create a new user account and send a verification email
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "User data" example({"email":"user@example.com","password":"password123","first_name":"John","last_name":"Doe"})
// @Success 201 {object} map[string]string "User created successfully" example({"message":"User created and verification mail sent"})
// @Failure 400 {object} map[string]string "Bad request" example({"message":"Invalid request data"})
// @Failure 500 {object} map[string]string "Internal server error" example({"message":"Could not save user"})
// @Router /signup [post]
func (h *UserHandler) Signup(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data!"})
		return
	}

	existingUser, err := h.userRepo.GetByEmail(c.Request.Context(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not check email", "error": err.Error()})
		return
	}

	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"message": "Email already taken"})
		return
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Role = "user"
	user.IsActive = true
	user.EmailVerified = false

	if err := h.userRepo.Create(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save user", "error": err.Error()})
		return
	}

	if err := h.sendVerify(user.Email, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not send verification email", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created and verification mail sent"})
}

// @Summary Log in a user
// @Description Authenticate a user and return a JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "User credentials" example({"email":"user@example.com","password":"password123"})
// @Success 200 {object} map[string]string "Login successful" example({"message":"login successful","token":"jwt-token-example"})
// @Failure 400 {object} map[string]string "Bad request" example({"message":"Invalid request data"})
// @Failure 401 {object} map[string]string "Unauthorized" example({"message":"Invalid credentials"})
// @Failure 500 {object} map[string]string "Internal server error" example({"message":"Could not authenticate user"})
// @Router /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data!"})
		return
	}

	validatedUser, err := h.userRepo.ValidateCredentials(c.Request.Context(), user.Email, user.Password)
	if err != nil {
		log.Println("Error validating credentials:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Could not authenticate user", "error": err.Error()})
		return
	}

	token, err := utils.GenerateToken(validatedUser.Email, validatedUser.ID)
	if err != nil {
		log.Println("Error generating token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not authenticate user", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "login successful", "token": token})
}

// @Summary Verify email
// @Description Verify a user's email using a token
// @Tags Auth
// @Accept json
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} map[string]string "Email verified successfully" example({"message":"Email verified successfully"})
// @Failure 400 {object} map[string]string "Bad request" example({"message":"Token is required"})
// @Failure 401 {object} map[string]string "Unauthorized" example({"message":"Invalid or expired token"})
// @Failure 404 {object} map[string]string "Not found" example({"message":"User not found"})
// @Failure 500 {object} map[string]string "Internal server error" example({"message":"Could not fetch user"})
// @Router /verify [get]
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	token := c.DefaultQuery("token", "")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Token is required"})
		return
	}

	userID, err := utils.VerifyToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
		return
	}

	verifiedUser, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not fetch user"})
		return
	}

	if verifiedUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if verifiedUser.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already verified"})
		return
	}

	if err := h.userRepo.UpdateEmailVerified(c.Request.Context(), verifiedUser.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update email verification status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *UserHandler) sendVerify(email string, userId int64) error {
	token, err := utils.GenerateVerifyToken(userId)
	if err != nil {
		return fmt.Errorf("error generating verification token: %v", err)
	}

	verifyURL := fmt.Sprintf("http://localhost:8080/verify?token=%s", token)
	body := fmt.Sprintf("Click to verify your email: %s", verifyURL)
	subject := "Verify Your Email"

	return h.emailService.SendEmail(email, subject, body)
}

// @Summary Get current user
// @Description Get the authenticated user's information
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID type"})
		return
	}
	if userID == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Update user details
// @Description Update the authenticated user's first name and last name
// @Tags User
// @Accept json
// @Produce json
// @Param user body map[string]string true "Updated user data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /me [put]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID type"})
		return
	}

	var updateData struct {
		FirstName string `json:"firstName" binding:"omitempty"`
		LastName  string `json:"lastName" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user", "error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if updateData.FirstName != "" {
		user.FirstName = updateData.FirstName
	}

	if updateData.LastName != "" {
		user.LastName = updateData.LastName
	}

	user.UpdatedAt = time.Now()
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update user", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": user})
}

// @Summary Change password
// @Description Change the authenticated user's password
// @Tags User
// @Accept json
// @Produce json
// @Param password body map[string]string true "Old and new passwords"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /change-password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID type"})
		return
	}

	var request struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user", "error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if !utils.CheckPasswordHash(request.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Old password is incorrect"})
		return
	}

	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not hash password", "error": err.Error()})
		return
	}

	if err := h.userRepo.UpdatePassword(c.Request.Context(), user.ID, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update password", "error": err.Error()})
		return
	}

	subject := "Password Updated Successfully"
	body := "Your password has been updated successfully. If you did not perform this action, please contact support immediately."
	mailer := models.NewMailer()
	if err := mailer.Mailer(user.Email, subject, body); err != nil {
		log.Println("Failed to send password update notification email:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// @Summary Get all users
// @Description Retrieve a list of all users (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {array} models.User "List of users"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID type"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user"})
		return
	}

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}

	users, err := h.userRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve users", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// @Summary Request password reset
// @Description Send a password reset email to the user
// @Tags Auth
// @Accept json
// @Produce json
// @Param email body map[string]string true "User email"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /forgot-password [post]
func (h *UserHandler) ForgetPassword(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(c.Request.Context(), request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not check email", "error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User with this email does not exist"})
		return
	}

	resetToken, err := utils.GenerateResetToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate reset token", "error": err.Error()})
		return
	}

	resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", resetToken)
	body := fmt.Sprintf("Click the link to reset your password: %s", resetURL)
	subject := "Password Reset Request"
	mailer := models.NewMailer()
	if err := mailer.Mailer(user.Email, subject, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not send reset email", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

// @Summary Reset password
// @Description Reset a user's password using a token
// @Tags Auth
// @Accept json
// @Produce json
// @Param reset body map[string]string true "Reset token and new password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var request struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	userID, err := utils.VerifyToken(request.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token", "error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user", "error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not hash password", "error": err.Error()})
		return
	}

	if err := h.userRepo.UpdatePassword(c.Request.Context(), user.ID, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update password", "error": err.Error()})
		return
	}

	subject := "Password Updated Successfully"
	body := "Your password has been updated successfully. If you did not perform this action, please contact support immediately."
	mailer := models.NewMailer()
	if err := mailer.Mailer(user.Email, subject, body); err != nil {
		log.Println("Failed to send password update notification email:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
