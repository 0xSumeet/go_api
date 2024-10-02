package auth

import (
	"net/http"
	"strings"

	"github.com/0xSumeet/go_api/internal/configs"
	"github.com/0xSumeet/go_api/pkg/utils"

	jwtlib "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the request header
		tokenString := c.GetHeader("Authorization")

		// Check if the token is provided and starts with "Bearer "
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(
				http.StatusUnauthorized,
				map[string]any{"error": "No or invalid authorization token provided"},
			)
			c.Abort()
			return
		}

		// Remove the "Bearer " prefix from the token string
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		// Parse the token
		claims := &utils.Claims{}
		token, err := jwtlib.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwtlib.Token) (interface{}, error) {
				return []byte(config.JWTSecret), nil // Replace with your secret key
			},
		)

		// Check if the token is valid
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, map[string]any{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// If the token is valid, store user info in the context
		c.Set("username", claims.Username)

		// Proceed with the request
		c.Next()
	}
}
