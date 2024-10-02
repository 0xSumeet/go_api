package main

import (
	"github.com/0xSumeet/go_api/internal/database"
	"github.com/0xSumeet/go_api/internal/routes"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

func main() {
	database.Init()
	// Close the db connection, after main function is executed
	defer database.DB.Close()

	app := gin.Default()
	routes.SetupRoutes(app)
	app.Run(":4000")
}
