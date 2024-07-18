package middlewares

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "strings"
    "ms-tts-go/utils"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
            c.Abort()
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
            c.Abort()
            return
        }

        token := bearerToken[1]
        if !utils.ValidateToken(token) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Next()
    }
}
