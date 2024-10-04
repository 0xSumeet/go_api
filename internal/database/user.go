package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"` // created_at
	UpdatedAt time.Time `json:"-"` // updated_at
}

type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Check if email field already exist in the database
func CheckIfEmailExists(user User) (bool, error) {
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

// Insert User to the database
func InsertUser(user User) (User, error) {
	var err error
	query := "INSERT INTO users (email, name, password) VALUES ($1, $2, $3)"
	_, err = DB.Exec(query, user.Email, user.Name, user.Password)
	if err != nil {
		return User{}, fmt.Errorf("could not create user")
	}
	return user, nil
}

// Insert User to the database
func CreateUser(user *User) (*User, error) {
	var err error
	query := "INSERT INTO users (email, name, password) VALUES ($1, $2, $3)"
	_, err = DB.Exec(query, user.Email, user.Name, user.Password)
	if err != nil {
		return &User{}, fmt.Errorf("could not create user")
	}

	newUserResponse := &User{
		Email: user.Email,
		Name:  user.Name,
	}
	return newUserResponse, nil
}

// Fetch password hash from the database
func FetchPasswordHash(user User) (string, error) {
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

func CreateNewUser(user *User) (*User, error) {
	var err error
	query := "INSERT INTO users (email, name, password) VALUES ($1, $2, $3) RETURNING id"
	err = DB.QueryRow(query, user.Email, user.Name, user.Password).Scan(&user.ID)
	if err != nil {
		return &User{}, fmt.Errorf("could not create user")
	}

	newUserResponse := &User{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}
	return newUserResponse, nil
}

func UpdateUser(user *User) (*User, error) {
	var err error
	updateQuery := `UPDATE users
        SET email = COALESCE($1, email),
            name = COALESCE($2, name),
            password = COALESCE($3, password),
            updated_at = NOW()
        WHERE id = $4;`

	_, err = DB.Exec(updateQuery, user.Email, user.Name, user.ID)
	if err != nil {
		return nil, fmt.Errorf("Error Updating fields: %v", err)
	}

	// Fetch the updated account result
	var updatedUser User
	fetchQuery := "SELECT id, email, name, updated_at FROM users WHERE id = $1;"

	// Execute the select query
	err = DB.QueryRow(fetchQuery, user.ID).
		Scan(&updatedUser.ID, &updatedUser.Email, &updatedUser.Name, &updatedUser.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error fetching user details: %v", err)
	}

	// Return Updated User
	return &updatedUser, nil
}

