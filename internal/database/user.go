package database

import (
	"database/sql"
	"fmt"

	"github.com/0xSumeet/go_api/internal/models"

	_ "github.com/lib/pq"
)

type User struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Check if email field already exist in the database
func CheckIfEmailExists(user models.User) (bool, error) {
	var err error
	var exists bool
	// var user models.User

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)"
	err = DB.QueryRow(query, user.Email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Check if email or password field is empty
func CheckIfEmailOrPasswordFieldEmpty(user models.User) (bool, error) {
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

// Insert User to the database
func InsertUser(user models.User) (models.User, error) {
	var err error
	query := "INSERT INTO users (email, name, password) VALUES ($1, $2, $3)"
	_, err = DB.Exec(query, user.Email, user.Name, user.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("could not create user")
	}
	return user, nil
}

// Insert User to the database
func CreateUser(user *User) (*models.UserResponse, error) {
	var err error
	query := "INSERT INTO users (email, name, password) VALUES ($1, $2, $3)"
	_, err = DB.Exec(query, user.Email, user.Name, user.Password)
	if err != nil {
		return &models.UserResponse{}, fmt.Errorf("could not create user")
	}

	newUserResponse := models.UserResponse{
		Email: user.Email,
		Name:  user.Name,
	}
	return &newUserResponse, nil
}

// Fetch password hash from the database
func FetchPasswordHash(user models.User) (string, error) {
	var storedHashedPassword string
	err := DB.QueryRow("SELECT password FROM users where email=$1", user.Email).
		Scan(&storedHashedPassword)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("invalid email or password")
	} else if err != nil {
		return "", fmt.Errorf("database error")
	}
	return storedHashedPassword, nil
}

