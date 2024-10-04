package utils

import (
	"fmt"

	"github.com/0xSumeet/go_api/internal/configs"
	"github.com/0xSumeet/go_api/internal/database"
)

type UserRequest struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate page number for request
func ValidatePageNumber(page int) (bool, error) {
	if page < 1 {
		return false, fmt.Errorf("Invalid Page number")
	}
	return true, nil
}

// Validate data limit for pagination
func ValidateDataLimit(limit int) int {
	if limit <= 0 {
		return config.DefaultLimit
	} else if limit > config.MaximumLimit {
		return config.MaximumLimit
	}
	return limit
}

// Check if email or password field is empty
func CheckIfEmailOrPasswordFieldEmpty(user database.User) (bool, error) {
	// Check if email is empty
	if user.Email == "" {
		return true, fmt.Errorf("please provide the email")
	}
	// Check if password is empty
	if user.Password == "" {
		return true, fmt.Errorf("please provide the password")
	}
	return false, nil
}

func CheckIfEmailOrPasswordFieldEmptyTry(userRequest *UserRequest) (bool, error) {
	// Check if email is empty
	if userRequest.Email == "" {
		return true, fmt.Errorf("please provide the email")
	}
	// Check if name is empty
	if userRequest.Name == "" {
		return true, fmt.Errorf("please provide the name")
	}
	// Check if password is empty
	if userRequest.Password == "" {
		return true, fmt.Errorf("please provide the password")
	}
	return false, nil
}

func CheckProductFields(product database.Product) (bool, error) {

	if product.ProductName == "" {
		return false, fmt.Errorf("error: product name cannot be empty")
	}

	if product.Price <= 0 {
		return false, fmt.Errorf("error: price cannot be a negative value")
	}

  if product.StockQuantity <= 0 {
    return false, fmt.Errorf("error: stock quantity cannot be negative value")
  }

	return true, nil
}
