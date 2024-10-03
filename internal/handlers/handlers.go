package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/0xSumeet/go_api/internal/database"
	"github.com/0xSumeet/go_api/internal/models"
	"github.com/0xSumeet/go_api/pkg/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func Home(c *gin.Context) {
	c.JSON(200, map[string]any{
		"message": "home page",
	})
}

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

func GetProductById(c *gin.Context) {
	// Get the 'id' parameter from the URL and convert it to an integer
	idParam := c.Param("id")

	// convert string to int
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid product ID"})
		return
	}

	// Query the database
	queryResult, err := database.GetProductByID(id)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"message": "error getting product", "error": err.Error()},
		)
		return
	}

	// Count the total product for headers
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

func SignUp(c *gin.Context) {
	/*
		type userRequest struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}

		var req userRequest
	*/
	var user models.User
	var err error

	// Bind JSON input to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	// Check if empty field
	_, err = database.CheckIfEmailOrPasswordFieldEmpty(user)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": err.Error()},
		)
		return
	}

	// Check if email exist
	userExist, err := database.CheckIfEmailExists(user)
	if userExist {
		message := fmt.Sprintf("this email alreadyy exists")
		c.JSON(http.StatusConflict, map[string]any{"error": message})
		return
	}

	// Password Hashing
	hashedPassword, err := utils.GenerateHash(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Could not hash password"})
		return
	}

	// Store the password hash in user.Password field
	user.Password = hashedPassword

	// Insert New User
	_, err = database.InsertUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// Respond with success
	userCreated := fmt.Sprintf("%s created successfully", user.Name)
	c.JSON(http.StatusCreated, map[string]any{"message": userCreated, "status": "success"})
}

func Login(c *gin.Context) {
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

	// Query the database for the stored password and
	// fetch the hash password
	hashPassword, err := database.FetchPasswordHash(user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}

	userPassword := user.Password

	// Compare the user input password with the fetched password
	_, err = utils.CompareHashPasswords(hashPassword, userPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": "invlaid password"})
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

	_, err = utils.ValidatePageNumber(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "error converting limit to int"})
	}

	// Validate data limit, by default, the return data limit is 10 and max limit is 20
	_ = utils.ValidateDataLimit(limit)

	products, err := database.PaginateData(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Cannot paginate data"})
		return
	}

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
