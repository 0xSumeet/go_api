package routes

import (
	"github.com/0xSumeet/go_api/internal/handlers"
	"github.com/0xSumeet/go_api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(c *gin.Engine) {
	c.GET("/home", handlers.Home)
	c.POST("/signup", handlers.SignUpTry)
	c.POST("/login", handlers.Login)
	c.POST("/register-product", handlers.AddProduct)
	c.PUT("/update-product/:id", handlers.UpdateProduct)

	// Auth Protected routes
	authorized := c.Group("/secure", auth.AuthMiddleware())
	{
		authorized.GET("/products", handlers.GetProductsByLimit)
		authorized.GET("/product/:id", handlers.GetProductById)
	}

	// c.GET("/users", handlers.GetUsers)
}
