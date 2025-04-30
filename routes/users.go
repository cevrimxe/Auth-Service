package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cevrimxe/auth-service/models"
	"github.com/cevrimxe/auth-service/utils"
	"github.com/gin-gonic/gin"
)

func signup(context *gin.Context) {
	var user models.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data!"})
		return
	}

	// is email unique?
	existingUser, err := models.GetUserByEmail(user.Email)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not check email", "error": err.Error()})
		return
	}

	if existingUser != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Email already taken"})
		return
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Role = "user"
	err = user.Save()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save!", "err": err})
		return
	}
	err = sendVerify(user.Email, user.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not send verification email", "err": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created and verification mail sent"})
}

func login(context *gin.Context) {
	var user models.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("Error binding JSON:", err) // Log hatayı
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data!"})
		return
	}

	err = user.ValidateCredentials()
	if err != nil {
		log.Println("Error validating credentials:", err) // Log hatayı
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Couldnt authenticate user", "error": err.Error()})
		return
	}

	token, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		log.Println("Error generating token:", err) // Log hatayı
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Couldnt authenticate user", "error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "login successful", "token": token})
}

func verifyEmail(context *gin.Context) {
	token := context.DefaultQuery("token", "")
	if token == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Token is required"})
		return
	}

	userID, err := utils.VerifyToken(token)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
		return
	}
	fmt.Println("User ID from token:", userID)

	verifiedUser, err := models.GetUserById(userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not fetch user"})
		return
	}

	if verifiedUser == nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if verifiedUser.EmailVerified {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Email already verified"})
		return
	}

	err = models.UpdateEmailVerified(verifiedUser.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update email verification status"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})

}

func sendVerify(email string, userId int64) error {
	fmt.Println("Generating token with userId:", userId)
	token, err := utils.GenerateVerifyToken(userId)
	if err != nil {
		return fmt.Errorf("error generating verification token: %v", err)
	}

	verifyURL := fmt.Sprintf("http://localhost:8080/verify?token=%s", token)
	body := fmt.Sprintf("Click to verify your email: %s", verifyURL)
	subject := "Verify Your Email"
	mailer := models.NewMailer()
	err = mailer.Mailer(email, subject, body)
	if err != nil {
		return err
	}
	return nil
}

func getMe(context *gin.Context) {
	userIDAny, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID type"})
		return
	}
	if !ok || userID == 0 {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid or missing user ID"})
		return
	}

	user, err := models.GetUserById(userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user"})
		return
	}

	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	context.JSON(http.StatusOK, user)
}
func forgetPassword(context *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	// Parse the request body
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	// Check if the user exists
	user, err := models.GetUserByEmail(request.Email)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not check email", "error": err.Error()})
		return
	}

	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "User with this email does not exist"})
		return
	}

	// Generate a password reset token
	resetToken, err := utils.GenerateResetToken(user.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate reset token", "error": err.Error()})
		return
	}

	// Send the reset token via email
	resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", resetToken)
	body := fmt.Sprintf("Click the link to reset your password: %s", resetURL)
	subject := "Password Reset Request"
	mailer := models.NewMailer()
	if err := mailer.Mailer(user.Email, subject, body); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not send reset email", "error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

func resetPassword(context *gin.Context) {
	var request struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}

	// Parse the request body
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	// Verify the reset token
	userID, err := utils.VerifyToken(request.Token)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token", "error": err.Error()})
		return
	}

	// Fetch the user by ID
	user, err := models.GetUserById(userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user", "error": err.Error()})
		return
	}

	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	err = user.UpdatePassword(request.NewPassword)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update password", "error": err.Error()})
		return
	}

	// Notify the user via email about the password update
	subject := "Password Updated Successfully"
	body := "Your password has been updated successfully. If you did not perform this action, please contact support immediately."
	mailer := models.NewMailer()
	if err := mailer.Mailer(user.Email, subject, body); err != nil {
		log.Println("Failed to send password update notification email:", err)
	}

	context.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func updateMe(context *gin.Context) {
	userIDAny, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID type"})
		return
	}

	var updateData struct {
		FirstName string `json:"firstName" binding:"omitempty"`
		LastName  string `json:"lastName" binding:"omitempty"`
	}

	if err := context.ShouldBindJSON(&updateData); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	user, err := models.GetUserById(userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user", "error": err.Error()})
		return
	}

	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if updateData.FirstName != "" {
		user.FirstName = updateData.FirstName
	}

	if updateData.LastName != "" {
		user.LastName = updateData.LastName
	}

	user.UpdatedAt = time.Now()
	if err := user.Update(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update user", "error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": user})
}
func changePassword(context *gin.Context) {
	userIDAny, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID type"})
		return
	}

	var request struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}

	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data", "error": err.Error()})
		return
	}

	user, err := models.GetUserById(userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user", "error": err.Error()})
		return
	}

	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err := user.CheckPassword(request.OldPassword); err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Old password is incorrect"})
		return
	}

	if err := user.UpdatePassword(request.NewPassword); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update password", "error": err.Error()})
		return
	}

	// Notify the user via email about the password update
	subject := "Password Updated Successfully"
	body := "Your password has been updated successfully. If you did not perform this action, please contact support immediately."
	mailer := models.NewMailer()
	if err := mailer.Mailer(user.Email, subject, body); err != nil {
		log.Println("Failed to send password update notification email:", err)
	}

	context.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func getUsers(context *gin.Context) {
	userIDAny, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid user ID type"})
		return
	}

	user, err := models.GetUserById(userID)
	if err != nil || user == nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve user"})
		return
	}

	if user.Role != "admin" {
		context.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}

	users, err := models.GetAllUsers()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve users", "error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"users": users})
}

// func logout(context *gin.Context) {
//
// 	context.Set("userId", nil)

//
// 	context.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
// }
