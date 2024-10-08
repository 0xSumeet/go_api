package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() {
	dbInfo := "postgres://postgres:password@localhost:5100/ecomdb?sslmode=disable"

	var err error

	DB, err = sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatalf("Error Connecting DB: %s", err)
	}

	// Ping database, and confirming connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("Error pinging database: %s", err)
	}

	// Set the time zone to 'Asia/Kolkata' (IST)
	_, err = DB.Exec("SET TIME ZONE 'Asia/Kolkata'")
	if err != nil {
		log.Fatalf("Error setting time zone: %s", err)
	}

	fmt.Println("Database connected and time zone set to Asia/Kolkata!")
}
