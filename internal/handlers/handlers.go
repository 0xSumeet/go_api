package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/0xSumeet/go_api/internal/configs"
	"github.com/0xSumeet/go_api/internal/database"
	"github.com/0xSumeet/go_api/internal/models"
	"github.com/0xSumeet/go_api/pkg/utils"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func Home(c *gin.Context) {
	c.JSON(200, map[string]any{
		"message": "home page",
	})
}

/*
func GetProductByID(c *gin.Context) {
	// Get the 'id' parameter from the URL and convert it to an integer
	idParam := c.Param("id")

	// convert string to int
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid product ID"})
		return
	}

	// Query the database for a product with the given ID
	var product models.Product
	query := "SELECT product_id, product_name, category, stock_quantity, price FROM products where product_id=$1"
	err = database.DB.QueryRow(query, id).
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
	err = database.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
		return
	}

	c.Header("X-Total-Count", strconv.Itoa(totalProducts))
	// Return the product as JSON
	c.JSON(http.StatusOK, product)
}
*/

func GetProductsPaginated(c *gin.Context) {
	// Get page number from query parameters (defaults to 1 if not provided)
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid page number"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > config.MaximumLimit {
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

	//func PaginateData(pagenumber, limit) ([]models.Product, error)
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

	// total User count
	var totalProducts int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
		return
	}

	// Query to get paginated products
	rows, err := database.DB.Query(
		"SELECT product_id, product_name, category, stock_quantity, price FROM products LIMIT $1 OFFSET $2",
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []models.Product

	// Loop through the rows and append each product to the slice
	for rows.Next() {
		var product models.Product
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

	c.Header("X-Current-Page", strconv.Itoa(page))
	c.Header("X-Offset", strconv.Itoa(offset))
	c.Header("X-Limit", strconv.Itoa(limit))
	c.Header("X-Default-Limit", strconv.Itoa(config.DefaultLimit))
	c.Header("X-Maximum-Limit", strconv.Itoa(config.MaximumLimit))
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

func SignUpFristTry(c *gin.Context) {
	var user models.User

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
	err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", user.Email).
		Scan(&exists)
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
	_, err = database.DB.Exec(
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
	token, err := utils.GenerateJWT(user.Name)
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

func Login(c *gin.Context) {
	var user models.User
	var err error

	if err = c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid Request"})
		return
	}

	// Query the database for the stored password
	var storedHashedPassword string

	err = database.DB.QueryRow("SELECT password FROM users where email=$1", user.Email).
		Scan(&storedHashedPassword)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, map[string]any{"message": "Invalid email or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"message": "database error"})
		return
	}

	// Compare the password
	if err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(user.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"message": "invalid password"})
		return
	}

	// Generate the JWT Token
	token, err := utils.GenerateJWT(user.Name)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": "Failed to generate JWT token"},
		)
		return
	}

	// Send the token as response
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged in",
		"token":   token,
		"status":  "success",
	})
}

func GetProductById(c *gin.Context) {
	// Get the 'id' parameter from the URL and convert it to an integer
	idParam := c.Param("id")

	// convert string to int
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid product ID"})
		return
	}
	queryResult, err := database.GetProductByID(id)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"message": "error getting product", "error": err.Error()},
		)
		return
	}
	/*
		// Query the database for a product with the given ID
		var product models.Product
		query := "SELECT product_id, product_name, category, stock_quantity, price FROM products where product_id=$1"
		err = database.DB.QueryRow(query, id).
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
	*/

	/*
		var totalProducts int
		err = database.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
			return
		}

		c.Header("X-Total-Count", strconv.Itoa(totalProducts))
		// Return the product as JSON
		c.JSON(http.StatusOK, models.Product)
	*/
	count, err := database.GetTotalProductsCount()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": "error counting products", "error": err.Error()},
		)
		return
	}
	// Return the total product count
	c.Header("X-Total-Count", strconv.Itoa(count))

	// Return the product in JSON
	c.JSON(http.StatusOK, queryResult)
}

/*
func GetProductsPaginated(c *gin.Context) {
	// Get page number from query parameters (defaults to 1 if not provided)
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid page number"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > config.MaximumLimit {
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

	// Define the number of items per page
	offset := (page - 1) * limit

	// total User count
	var totalProducts int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
		return
	}

	// Query to get paginated products
	rows, err := database.DB.Query(
		"SELECT product_id, product_name, category, stock_quantity, price FROM products LIMIT $1 OFFSET $2",
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []models.Product

	// Loop through the rows and append each product to the slice
	for rows.Next() {
		var product models.Product
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

	c.Header("X-Current-Page", strconv.Itoa(page))
	c.Header("X-Offset", strconv.Itoa(offset))
	c.Header("X-Limit", strconv.Itoa(limit))
	c.Header("X-Default-Limit", strconv.Itoa(config.DefaultLimit))
	c.Header("X-Maximum-Limit", strconv.Itoa(config.MaximumLimit))
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
*/

func GetProducts(c *gin.Context) {
	// Run the query function GetProducts()
	queryResult, err := database.GetProducts()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"message": "error getting products", "error": err.Error()},
		)
		return
	}
	// total product count in header
	count, err := database.GetTotalProductsCount()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"message": "error counting products", "error": err.Error()},
		)
		return
	}
	// header
	c.Header("X-Total-Products-Count", strconv.Itoa(count))

	// Return the product in JSON
	c.JSON(http.StatusOK, queryResult)
}

func SignUpSecondTry(c *gin.Context) {
	var user models.User
	var err error

	// Bind JSON input to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	/*
		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not hash password"})
			return
		}
	*/

	/*
	  // Password Hashing
	  hashedPassword, err := utils.GenerateHash(user.Password)
	  if err != nil {
	    c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not hash password"})
	  }

	  user.Password = hashedPassword
	*/

	/*
		// Check if email exists
		var exists bool
		err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", user.Email).
			Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": "Database error"})
			return
		}

		if exists {
			existMessage := fmt.Sprintf("%s already exist", user.Email)
			c.JSON(http.StatusConflict, map[string]any{"error": existMessage})
			return
		}
	*/

	// Check if empty field
	_, err = database.CheckIfEmailOrPasswordFieldEmpty(user)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": err.Error()},
		)
		return
	}
	/*
		if isEmpty {
			c.JSON(
				http.StatusBadRequest,
				map[string]any{
					"message": "email or password field cannot be empty",
					"error":   err.Error(),
				},
			)
			return
		}
	*/

	// Check if email exist
	userExist, err := database.CheckIfEmailExists(user)
	/*
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				map[string]any{"error": "error checking user existence"},
			)
			//return
		}
	*/

	if userExist {
		message := fmt.Sprintf("%s already exists", user.Email)
		c.JSON(http.StatusConflict, map[string]any{"error": message})
		return
	}

	// Check if empty password

	// Password Hashing
	hashedPassword, err := utils.GenerateHash(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not hash password"})
		return
	}

	user.Password = hashedPassword

	/*
		// Insert new user into the database
		_, err = database.DB.Exec(
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
	*/

	// Insert New User
	_, err = database.InsertUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	// Respond with success
	userCreated := fmt.Sprintf("%s created successfully", user.Name)
	c.JSON(http.StatusCreated, map[string]any{"message": userCreated})

	// Generate the JWT Token
	token, err := utils.GenerateJWT(user.Name)
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

func SignUp(c *gin.Context) {
	var user models.User
	var err error

	// Bind JSON input to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	/*
		// Password Hashing
		hashedPassword, err := utils.GenerateHash(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not hash password"})
		}

		user.Password = hashedPassword

	*/
	/*
		// Check if email exists
		var exists bool
		err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", user.Email).
			Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": "Database error"})
			return
		}

		if exists {
			existMessage := fmt.Sprintf("%s already exist", user.Email)
			c.JSON(http.StatusConflict, map[string]any{"error": existMessage})
			return
		}
	*/

	// Check if empty field
	_, err = database.CheckIfEmailOrPasswordFieldEmpty(user)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": err.Error()},
		)
		return
	}
	/*
		if isEmpty {
			c.JSON(
				http.StatusBadRequest,
				map[string]any{
					"message": "email or password field cannot be empty",
					"error":   err.Error(),
				},
			)
			return
		}
	*/

	// Check if email exist
	userExist, err := database.CheckIfEmailExists(user)
	/*
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				map[string]any{"error": "error checking user existence"},
			)
			//return
		}
	*/

	if userExist {
		message := fmt.Sprintf("this email alreadyy exists")
		c.JSON(http.StatusConflict, map[string]any{"error": message})
		return
	}

	// Check if empty password

	// Password Hashing
	hashedPassword, err := utils.GenerateHash(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not hash password"})
		return
	}

	user.Password = hashedPassword

	/*
		// Insert new user into the database
		_, err = database.DB.Exec(
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
	*/

	// Insert New User
	_, err = database.InsertUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// Respond with success
	userCreated := fmt.Sprintf("%s created successfully", user.Name)
	c.JSON(http.StatusCreated, map[string]any{"message": userCreated, "status": "success"})

	/*
		// Generate the JWT Token
		token, err := utils.GenerateJWT(user.Name)
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
	*/
}

func LoginTry(c *gin.Context) {
	var user models.User
	//	var err error

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid Request"})
		return
	}

	// Check if empty field
	_, err := database.CheckIfEmailOrPasswordFieldEmpty(user)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": err.Error()},
		)
		return
	}
	// Query the database for the stored password
	// var storedHashedPassword string

	// Fetch the hash password
	hashPassword, err := database.FetchPasswordHash(user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}

	userPassword := user.Password

	// Compare the user input password with the fetched password
	_, err = database.CompareHashPasswords(hashPassword, userPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": "invlaid password"})
		return
	}

	/*
		  err := database.DB.QueryRow("SELECT password FROM users where email=$1", user.Email).
				Scan(&storedHashedPassword)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, map[string]any{"message": "Invalid email or password"})
				return
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"message": "database error"})
				return
			}
	*/
	/*
		// Compare the password
		if err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(user.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, map[string]any{"message": "invalid password"})
			return
		}
	*/

	// Generate the JWT Token
	token, err := utils.GenerateJWT(user.Name)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": "Failed to generate JWT token"},
		)
		return
	}

	// Send the token as response
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged in",
		"token":   token,
		"status":  "success",
	})
}

/*
func PaginateData(pagenumber, limit) ([]models.Product, error) {

}
*/

func GetProductsByLimit(c *gin.Context) {
	// Get page number from query parameters (defaults to 1 if not provided)
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			map[string]any{"error": "error converting page number to int"},
		)
		return
	}
	/*
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid page number"})
			return
	*/
	_, err = database.ValidatePageNumber(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "error converting limit to int"})
	}

  //Validate data limit, by default, the return data limit is 10 and max limit is 20
	_ = database.ValidateDataLimit(limit)

	/*
		if err != nil || limit > config.MaximumLimit {
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

		// Define the number of items per page
		offset := (page - 1) * limit

		// total User count
		var totalProducts int
		err = database.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not count users"})
			return
		}
	*/
	/*
		// Query to get paginated products
		rows, err := database.DB.Query(
			"SELECT product_id, product_name, category, stock_quantity, price FROM products LIMIT $1 OFFSET $2",
			limit,
			offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		defer rows.Close()

		var products []models.Product

		// Loop through the rows and append each product to the slice
		for rows.Next() {
			var product models.Product
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
	*/
	products, err := database.PaginateData(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Cannot paginate data"})
		return
	}
	/*
		c.Header("X-Current-Page", strconv.Itoa(page))
		c.Header("X-Offset", strconv.Itoa(offset))
		c.Header("X-Limit", strconv.Itoa(limit))
		c.Header("X-Default-Limit", strconv.Itoa(config.DefaultLimit))
		c.Header("X-Maximum-Limit", strconv.Itoa(config.MaximumLimit))
		c.Header("X-Product-Count", strconv.Itoa(len(products)))
		c.Header("X-Total-Count", strconv.Itoa(totalProducts))
		c.Header("Total-Pages", strconv.Itoa(totalProducts/limit))

	*/
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
