package models

/*
type User struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
*/

type UserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
