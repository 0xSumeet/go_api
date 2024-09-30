package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/0xSumeet/go_api/jwt"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	// Default Limit and Max limit of the data to be fetched
	defaultLimit int = 5
	maximumLimit int = 20
)

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

func init() {
	dbInfo := "postgres://postgres:password@localhost:5100/ecomdb?sslmode=disable"

	var err error

	db, err = sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatalf("Error Connecting DB: %s", err)
	}

	// Ping database, and confirming connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging db: %s", err)
	}
	fmt.Println("Database connected!")
}

func main() {
	defer db.Close()
	router := gin.Default()
	router.GET("/", home)

	router.POST("/signup", signUp)
	router.POST("/login", login)

	authorized := router.Group("/sec", jwt.AuthMiddleware())
	{
		authorized.GET("/products", getProductsPaginated)
		authorized.GET("/product/:id", retriveProductByID)
	}
	router.Run(":4000")
}

func home(c *gin.Context) {
	c.JSON(200, map[string]any{
		"message": "home page",
	})
}

func retriveProductByID(c *gin.Context) {
	// Get the 'id' parameter from the URL and convert it to an integer
	idParam := c.Param("id")

	// convert string to int
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid product ID"})
		return
	}

	// Query the database for a product with the given ID
	var product Product
	query := "SELECT product_id, product_name, category, stock_quantity, price FROM products where product_id=$1"
	err = db.QueryRow(query, id).
		Scan(&product.ID, &product.ProductName, &product.Category, &product.StockQuantity, &product.Price)

	if err == sql.ErrNoRows {
		c.JSON(
			http.StatusNotFound,
			map[string]any{
				"error":  "Product not found, Please enter valid product id",
				"status": "failure",
			},
		)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	var totalProducts int
	err = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
		return
	}

	c.Header("X-Total-Count", strconv.Itoa(totalProducts))
	// Return the product as JSON
	c.JSON(http.StatusOK, product)
}

func getProductsPaginated(c *gin.Context) {
	// Get page number from query parameters (defaults to 1 if not provided)
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid page number"})
		return
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > maximumLimit {
		c.JSON(
			http.StatusBadRequest,
			map[string]any{
				"error":  "limit value exceeded, please add small limit value",
				"status": "failure",
			},
		)
		return
	} else if limit < 1 {
		c.JSON(http.StatusNotFound,
			map[string]any{
				"error":  "No Data",
				"status": "failure",
			},
		)
		return
	}

	/*
		// Check limit
		if limitVal <= 0 {
			limitVal = defaultLimit
			return
		} else if limitVal > maximumLimit {
			limitVal = maximumLimit
			return
		}
	*/

	// Define the number of items per page
	offset := (page - 1) * limit

	/*
	   Use
	   http://localhost:4000/paginated?page=2 to get JSON result, the defualt value is 10 so the result will be from 10 to 20

	   You can also set the limit of how many result you want to obtain, by default, 10 result will obtained
	   http://localhost:4000/paginated?limit=2
	   http://localhost:4000/paginated?page=2&limit=20
	   http://localhost:4000/paginated?offset=0&limit=10
	*/
	// total User count
	var totalProducts int
	err = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
		return
	}

	// Query to get paginated products
	rows, err := db.Query(
		"SELECT product_id, product_name, category, stock_quantity, price FROM products LIMIT $1 OFFSET $2",
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []Product

	// Loop through the rows and append each product to the slice
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.ProductName, &product.Category, &product.StockQuantity, &product.Price); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		products = append(products, product)
	}

	// Check for errors during row iteration
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	/* if no product are found, return the result with following message
	if len(products) == 0 {
		c.JSON(http.StatusOK, map[string]any{"message": "no product found for this page!"})
	}
	*/
	c.Header("X-Current-Page", strconv.Itoa(page))
	c.Header("X-Offset", strconv.Itoa(offset))
	c.Header("X-Limit", strconv.Itoa(limit))
	c.Header("X-Default-Limit", strconv.Itoa(defaultLimit))
	c.Header("X-Maximum-Limit", strconv.Itoa(maximumLimit))
	c.Header("X-Product-Count", strconv.Itoa(len(products)))
	c.Header("X-Total-Count", strconv.Itoa(totalProducts))
	c.Header("Total-Pages", strconv.Itoa(totalProducts/limit))

	// Return products as JSON
	c.JSON(http.StatusOK, products)
	if products == nil {
		c.JSON(
			http.StatusBadRequest,
			map[string]any{"message": "no products available", "status": "failure"},
		)
	} else {
		c.JSON(http.StatusOK, map[string]any{"message": "Products", "status": "success"})
	}
}

func signUp(c *gin.Context) {
	var user User

	// Bind JSON input to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not hash password"})
		return
	}

	// Check if email exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", user.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Database error"})
		return
	}

	if exists {
		existMessage := fmt.Sprintf("%s already exist", user.Email)
		c.JSON(http.StatusConflict, map[string]any{"error": existMessage})
		return
	}

	// Insert new user into the database
	_, err = db.Exec(
		"INSERT INTO users (email, name, password) VALUES ($1, $2, $3)",
		user.Email,
		user.Name,
		string(hashedPassword),
	)
	if err != nil {
		userError := fmt.Sprintf("Could not create user %s", user.Name)
		c.JSON(http.StatusInternalServerError, map[string]any{"message": userError})
		return
	}

	// Respond with success
	userCreated := fmt.Sprintf("%s created successfully", user.Name)
	c.JSON(http.StatusCreated, map[string]any{"message": userCreated})

	// Generate the JWT Token
	token, err := jwt.GenerateJWT(user.Name)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": "Failed to generate JWT token"},
		)
		return
	}

	// Send the token as response
	c.JSON(http.StatusOK, map[string]any{
		"message": "success",
	})
	c.Header("Token", token)
}

func login(c *gin.Context) {
	var user User
	var err error

	if err = c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid Request"})
		return
	}
	//	fmt.Println("Email being queried:", user.Email)

	// Query the database for the stored password
	var storedHashedPassword string

	err = db.QueryRow("SELECT password FROM users where email=$1", user.Email).
		Scan(&storedHashedPassword)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, map[string]any{"message": "Invalid email or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"message": "database error"})
		return
	}
	// fmt.Println("Fetched hashed password:", storedHashedPassword)

	// Compare the password
	if err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(user.Password)); err != nil {
		// fmt.Println(err)
		c.JSON(http.StatusUnauthorized, map[string]any{"message": "invalid password"})
		return
	}

	// Generate the JWT Token
	token, err := jwt.GenerateJWT(user.Name)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": "Failed to generate JWT token"},
		)
		return
	}

	// Send the token as response
	c.JSON(http.StatusOK, map[string]any{
		"token": token,
	})
}
