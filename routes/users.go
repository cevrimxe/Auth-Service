package routes

import (
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
	context.JSON(http.StatusCreated, gin.H{"message": "User created"})

	mailer := models.NewMailer()
	err = mailer.SendVerifyEmail(user.Email, "denemetokeni")
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not send verification mail!"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Verification mail sent!"})
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
