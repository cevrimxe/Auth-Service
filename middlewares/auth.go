package middlewares

import (
	"net/http"

	"github.com/cevrimxe/auth-service/utils"
	"github.com/gin-gonic/gin"
)

func Authenticate(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")

	if token == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "not authorized token empty"})
		return
	}

	userId, err := utils.VerifyToken(token)

	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "not authorized"})

		return
	}

	context.Set("userId", userId)

	context.Next()
}
