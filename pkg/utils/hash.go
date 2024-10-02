package utils

import (
	"golang.org/x/crypto/bcrypt"
  "github.com/0xSumeet/go_api/internal/models"
)

func GenerateHash(password string) (string, error) {
  var user models.User
  var err error
  hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)  
  if err != nil {
    return "", err 
  }
  return string(hash), nil
}

//func CompareHash(models.User, dbPassword) bool {}

