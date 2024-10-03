package models

type User struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Product struct {
	ID            int     `json:"id"`
	ProductName   string  `json:"product_name"`
	Category      string  `json:"category"`
	StockQuantity int     `json:"stock_quantity"`
	Price         float64 `json:"price"`
}

type UserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
