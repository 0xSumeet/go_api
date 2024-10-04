package utils

import (
  "fmt"
	"golang.org/x/crypto/bcrypt"
  "github.com/0xSumeet/go_api/internal/database"
)

// Generate Password hash
func GenerateHash(password string) (string, error) {
  var user database.User
  var err error
  hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)  
  if err != nil {
    return "", err 
  }
  return string(hash), nil
}

// Compare user input password with the hash password from the database
func CompareHashPasswords(hash string, userpassword string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(userpassword)); err != nil {
		return false, fmt.Errorf("invalid password")
	}
	return true, nil
}



