package utils

import (
	"time"

	"github.com/0xSumeet/go_api/internal/configs"

	"github.com/dgrijalva/jwt-go"
)

// JWT Claims structure
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Generate token
func GenerateJWT(username string) (string, error) {
	// Define expiration time of the token
	expirationTime := time.Now().Add(5 * time.Minute)

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
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
