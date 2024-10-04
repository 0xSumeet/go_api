package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/0xSumeet/go_api/internal/database"
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

func UpdateProduct(c *gin.Context) {
	// Get the 'id' parameter from the URL and convert it to an integer
	idParam := c.Param("id")

	// Convert string to int
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid Product id"})
		return
	}

	var product database.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	_, err = utils.CheckProductFields(product)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"status": "failure", "error": err.Error()})
		return
	}

	product.ID = id
	updatedProduct, err := database.UpdateProductField(&product)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"message": "Could not update product", "error": err.Error()},
		)
		return
	}

	c.JSON(http.StatusOK, map[string]any{"message": "success", "data": updatedProduct})
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
	var user database.User
	var err error

	// Bind JSON input to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	// Check if empty field
	_, err = utils.CheckIfEmailOrPasswordFieldEmpty(user)
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
	fmt.Println(hashedPassword)
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
	var user database.User
	//	var err error

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid Request"})
		return
	}

	// Check if empty field
	_, err := utils.CheckIfEmailOrPasswordFieldEmpty(user)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			map[string]any{"error": err.Error()},
		)
		return
	}

	// Query the database for the stored password using FetchPasswordHash function and
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

type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

/*
type UserRequest struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
*/

func SignUpTry(c *gin.Context) {
	var userRequest struct {
		ID       int    `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	var user database.User
	var err error

	// Bind JSON input to the user struct
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	// Check if email empty, name or password field empty
	if userRequest.Email == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "please provide the email"})
		return
	}
	// Check if name is empty
	if userRequest.Name == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "please provide the name"})
		return
	}
	// Check if password is empty
	if userRequest.Password == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "please provide the password"})
		return
	}

	//	return false, nil
	/*
			// Check if empty field
		  _, err := utils.CheckIfEmailOrPasswordFieldEmptyTry(&utils.UserRequest)
			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					map[string]any{"error": err.Error()},
				)
				return
			}
	*/

	email := database.User{
		Email: userRequest.Email,
	}

	// Check if email exist
	emailExist, err := database.CheckIfEmailExists(email)
	if emailExist {
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

	response, err := database.CreateNewUser(&database.User{
		ID:       userRequest.ID,
		Name:     userRequest.Name,
		Email:    userRequest.Email,
		Password: user.Password,
	})
	/*
		// Insert New User
		_, err = database.InsertUser(user)
	*/if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	userResponse := UserResponse{
		ID:    response.ID,
		Email: response.Email,
		Name:  response.Name,
	}

	/*
		// Respond with success
		userCreated := fmt.Sprintf("%s created successfully", Response.Name)
		c.JSON(http.StatusCreated, map[string]any{"message": userCreated, "status": "success"})
	*/

	c.JSON(http.StatusCreated, map[string]any{"message": userResponse, "status": "success"})
}

func UpdateUser(c *gin.Context) {
	var userRequest struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	// Bind JSON input to the user struct
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	response, err := database.UpdateUser(&database.User{
		ID:    userRequest.ID,
		Email: userRequest.Email,
		Name:  userRequest.Name,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, map[string]any{"message": response, "status": "success"})
}

func AddProduct(c *gin.Context) {
	var product database.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	query, err := database.CreateProduct(&product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{"message": "success", "data": query})
}
