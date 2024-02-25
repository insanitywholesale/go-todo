package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite" // no-CGo database/sql driver for sqlite
)

// Global database object
var db *sql.DB

// SetupDB initializes the database and returns a client object to access it
func SetupDB() (*sql.DB, error) {
	// Get database file path from environment
	dbPath := os.Getenv("TODO_DB_PATH")
	// If the environment variable is empty use a default path of the current directory
	if dbPath == "" {
		dbPath = "todo.db"
	}

	// Check if file exists and if not, create it
	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		// Handle possible error types differently
		switch {
		case errors.Is(err, os.ErrPermission): // If there is a permission error just exit
			return nil, fmt.Errorf("[SetupDB] error with permissions trying to open file: %w", err)
		case errors.Is(err, os.ErrNotExist): // If the file doesn't exist try to create it
			_, err := os.Create(dbPath)
			if err != nil {
				return nil, fmt.Errorf("[SetupDB] error trying to create file: %w", err)
			}
		default: // For any other error return a generic error message
			return nil, fmt.Errorf("[SetupDB] error trying to open file: %w", err)
		}
	}

	// Ensure the file information is there and check if the file is a regular file
	if fileInfo != nil && !fileInfo.Mode().IsRegular() {
		return nil, errors.New("[SetupDB] error: " + dbPath + " is not a regular file")
	}

	// Since the file exists, use it for sqlite
	sqlite, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("[SetupDB] error opening sqlite file: %w", err)
	}

	// Ping the database to make sure we can access it and use it
	err = sqlite.Ping()
	if err != nil {
		return nil, fmt.Errorf("[SetupDB] error pinging sqlite database: %w", err)
	}

	// Create todos table
	_, err = sqlite.Exec(`CREATE TABLE if not exists todo (
		id INTEGER NOT NULL,
		description TEXT NOT NULL,
        done BOOLEAN NOT NULL DEFAULT(TRUE),
		PRIMARY KEY (id AUTOINCREMENT)
	);`)
	if err != nil {
		return nil, fmt.Errorf("[SetupDB] error creating todo database: %w", err)
	}

	return sqlite, nil
}

func main() {
	// Create database object
	mydb, err := SetupDB()
	if err != nil {
		log.Fatalln("[main] error setting up database:", err)
	}

	// Print a nice message on the terminal
	log.Println("[main] database intialized successfully")

	// Assign returned database client object to the global variable
	db = mydb

	// Check that the databse is accessible through the global variable as well
	err = db.Ping()
	if err != nil {
		log.Fatalln("[main] error pinging sqlite database:", err)
	}
}
