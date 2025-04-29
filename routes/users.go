package routes

import (
	"net/http"
	"time"

	"github.com/cevrimxe/auth-service/models"
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
}
