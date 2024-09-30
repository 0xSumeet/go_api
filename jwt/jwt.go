package jwt

import (
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtSecret = []byte("my_secret_key")

// JWT Claims structure
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Generate token
func GenerateJWT(username string) (string, error) {
	// Define expiration time of the token
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims, which includes the username and the expiry time
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create the token with the specified claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

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
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil // Replace with your secret key
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
