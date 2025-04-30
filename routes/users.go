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
	log.Println("Keys map:", context.Keys)
	userIDAny, exists := context.Get("userId")
	log.Println("useridany:", userIDAny, " exists:", exists)
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		log.Println("Unauthorized access attempt in getMe function")
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
