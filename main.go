package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite" // no-CGo database/sql driver for sqlite
)

// Global database object
var db *sql.DB

// HTTP handler for the root endpoint
func Home(w http.ResponseWriter, _ *http.Request) {
	welcomeMessage := "Welcome to the Todo API demo"
	_, err := w.Write([]byte(welcomeMessage))
	if err != nil {
		log.Println("[Home] error writing to client:", err)
	}
}

// HTTP handler for getting all todo items
func ReadTodos(w http.ResponseWriter, _ *http.Request) {
	// Get all rows from the todo table in the database
	rows, err := db.Query(`SELECT id, description, done FROM todo;`)
	if err != nil {
		// If the error is that no rows were returned that's not actually our problem
		if errors.Is(err, sql.ErrNoRows) {
			// Tell the client that we are going to return JSON
			w.Header().Add("Content-Type", "application/json")
			// Tell the client that the status of the request is 404
			w.WriteHeader(http.StatusNotFound)
			// Create a new error of our custom type
			e := NewHTTPError(err.Error(), http.StatusNotFound, "Not Found")
			// Return the JSON-encoded error message
			err := json.NewEncoder(w).Encode(e)
			if err != nil {
				// Log encoding error for debugging
				log.Println("[ReadTodos] error encoding error:", err)
			}
			return
		}
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 500
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[ReadTodos] error encoding error:", err)
		}
		return
	}

	// Store todo items in a slice
	var todos []*TodoItem

	// Leave closing the rows object for later after we've read the results
	defer rows.Close()
	// Map each returned row from the database query to a temp item and append to the slice
	for rows.Next() {
		// Check if there is an error without scanning
		if rows.Err() != nil {
			// Tell the client that we are going to return JSON
			w.Header().Add("Content-Type", "application/json")
			// Tell the client that the status of the request is 500
			w.WriteHeader(http.StatusInternalServerError)
			// Create a new error of our custom type
			e := NewHTTPError(rows.Err().Error(), http.StatusInternalServerError, "General Error")
			// Return the JSON-encoded error message
			err := json.NewEncoder(w).Encode(e)
			if err != nil {
				// Log encoding error for debugging
				log.Println("[ReadTodos] error encoding error:", err)
				return
			}
		}

		// Store current todo item we're working with
		var todo TodoItem

		// Put row from database into variable
		err = rows.Scan(&todo.ID, &todo.Description, &todo.Done)
		if err != nil {
			// If the error is that no rows were returned that's not actually our problem
			if errors.Is(err, sql.ErrNoRows) {
				// Tell the client that we are going to return JSON
				w.Header().Add("Content-Type", "application/json")
				// Tell the client that the status of the request is 404
				w.WriteHeader(http.StatusNotFound)
				// Create a new error of our custom type
				e := NewHTTPError(err.Error(), http.StatusNotFound, "Not Found")
				// Return the JSON-encoded error message
				err := json.NewEncoder(w).Encode(e)
				if err != nil {
					// Log encoding error for debugging
					log.Println("[ReadTodo] error encoding error:", err)
				}
				return
			}
			// Tell the client that we are going to return JSON
			w.Header().Add("Content-Type", "application/json")
			// Tell the client that the status of the request is 500
			w.WriteHeader(http.StatusInternalServerError)
			// Create a new error of our custom type
			e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
			// Return the JSON-encoded error message
			err := json.NewEncoder(w).Encode(e)
			if err != nil {
				// Log encoding error for debugging
				log.Println("[ReadTodos] error encoding error:", err)
				return
			}
		}

		// Add returned todo item to the list
		todos = append(todos, &todo)
	}

	// Tell the client that we are going to return JSON
	w.Header().Add("Content-Type", "application/json")
	// Tell the client that the status of the request is 200
	w.WriteHeader(http.StatusOK)
	// Return the JSON-encoded list of todo items
	err = json.NewEncoder(w).Encode(todos)
	if err != nil {
		// Log encoding error for debugging
		log.Println("[ReadTodos] error encoding todo item list:", err)
	}
}

// HTTP handler for getting a todo item
func ReadTodo(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter named todo_id
	todoIDfromURL := r.PathValue("todo_id")
	if todoIDfromURL == "" {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 400
		w.WriteHeader(http.StatusBadRequest)
		// Create a new error of our custom type
		e := NewHTTPError("Parameter todo_id is empty", http.StatusBadRequest, "Bad Request")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[ReadTodo] error encoding error:", err)
		}
		return
	}

	todoID, err := strconv.Atoi(todoIDfromURL)
	if err != nil {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 400
		w.WriteHeader(http.StatusBadRequest)
		// Create a new error of our custom type
		e := NewHTTPError("Parameter todo_id is not a number", http.StatusBadRequest, "Bad Request")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[ReadTodo] error encoding error:", err)
		}
		return
	}

	// Todo item we are going to return
	var todo TodoItem

	// Get row from the todo table in the database with the id
	row := db.QueryRow(`SELECT id, description, done FROM todo WHERE id = ?;`, todoID)
	// Put row from database into variable
	err = row.Scan(&todo.ID, &todo.Description, &todo.Done)
	if err != nil {
		// If the error is that no rows were returned that's not actually our problem
		if errors.Is(err, sql.ErrNoRows) {
			// Tell the client that we are going to return JSON
			w.Header().Add("Content-Type", "application/json")
			// Tell the client that the status of the request is 404
			w.WriteHeader(http.StatusNotFound)
			// Create a new error of our custom type
			e := NewHTTPError(err.Error(), http.StatusNotFound, "Not Found")
			// Return the JSON-encoded error message
			err := json.NewEncoder(w).Encode(e)
			if err != nil {
				// Log encoding error for debugging
				log.Println("[ReadTodo] error encoding error:", err)
			}
			return
		}
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 500
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[ReadTodo] error encoding error:", err)
			return
		}
	}

	// Tell the client that we are going to return JSON
	w.Header().Add("Content-Type", "application/json")
	// Tell the client that the status of the request is 200
	w.WriteHeader(http.StatusOK)
	// Return the JSON-encoded todo item
	err = json.NewEncoder(w).Encode(todo)
	if err != nil {
		// Log encoding error for debugging
		log.Println("[ReadTodo] error encoding todo item:", err)
	}
}

// HTTP handler for creating a todo item
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	// Todo item from the request body
	var todo TodoItem

	// Map todo from request body to variable
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 500
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		// Log encoding error for debugging
		log.Println("[CreateTodo] error encoding error:", err)
		return
	}

	// Save todo item in database and return generated id
	res, err := db.Exec(`INSERT INTO todo (done, description) VALUES (?, ?) RETURNING id;`,
		&todo.Done,
		&todo.Description,
	)
	if err != nil {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 500
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[CreateTodo] error encoding error:", err)
		}
		return
	}

	// Get id of last inserted row which should be the autoincrement id of the todo
	rowID, err := res.LastInsertId()
	if err != nil {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 500
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[CreateTodo] error encoding error:", err)
		}
		return
	}

	// Set todo ID to last insert ID
	todo.ID = rowID

	// Tell the client that we are going to return JSON
	w.Header().Add("Content-Type", "application/json")
	// Tell the client that the status of the request is 200
	w.WriteHeader(http.StatusOK)
	// Return the JSON-encoded new todo item
	err = json.NewEncoder(w).Encode(todo)
	if err != nil {
		// Log encoding error for debugging
		log.Println("[CreateTodo] error encoding todo item:", err)
		return
	}
}

// HTTP handler for updating a todo item
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter named todo_id
	todoIDfromURL := r.PathValue("todo_id")
	if todoIDfromURL == "" {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 400
		w.WriteHeader(http.StatusBadRequest)
		// Create a new error of our custom type
		e := NewHTTPError("Parameter todo_id is empty", http.StatusBadRequest, "Bad Request")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[UpdateTodo] error encoding error:", err)
		}
		return
	}

	// Convert string variable from URL to integer
	todoID, err := strconv.Atoi(todoIDfromURL)
	if err != nil {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 400
		w.WriteHeader(http.StatusBadRequest)
		// Create a new error of our custom type
		e := NewHTTPError("Parameter todo_id is not a number", http.StatusBadRequest, "Bad Request")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[UpdateTodo] error encoding error:", err)
		}
		return
	}

	// Todo item from the request body
	var todo TodoItem

	// Map todo from request body to variable
	err = json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 500
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[UpdateTodo] error encoding error:", err)
		}
		return
	}
	// Set its ID equal to the URL path variable
	todo.ID = int64(todoID)

	// Update todo item in database based on specified id
	res, err := db.Exec(`UPDATE todo SET done = ?, description = ? WHERE id = ?;`,
		&todo.Done,
		&todo.Description,
		todoID, // URL path variable instead of the ID in the struct
	)
	if err != nil {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 500
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[UpdateTodo] error encoding error:", err)
		}
		return
	}

	// Get amount of changed rows
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 500
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError(err.Error(), http.StatusInternalServerError, "General Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[UpdateTodo] error encoding error:", err)
		}
		return
	}

	switch rowsAffected {
	case 0:
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 404
		w.WriteHeader(http.StatusNotFound)
		// Create a new error of our custom type
		e := NewHTTPError("No todo with id "+strconv.Itoa(todoID)+" exists", http.StatusNotFound, "Not Found")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[DeleteTodo] error encoding error:", err)
		}
		return
	case 1:
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 200
		w.WriteHeader(http.StatusOK)
		// Return the JSON-encoded new todo item
		err = json.NewEncoder(w).Encode(todo)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[UpdateTodo] error encoding todo item:", err)
			return
		}
	default:
		// Tell the client that we are going to return JSON
		w.Header().Add("Content-Type", "application/json")
		// Tell the client that the status of the request is 404
		w.WriteHeader(http.StatusInternalServerError)
		// Create a new error of our custom type
		e := NewHTTPError("deleted more than 1 record", http.StatusInternalServerError, "Internal Server Error")
		// Return the JSON-encoded error message
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			// Log encoding error for debugging
			log.Println("[UpdateTodo] error encoding error:", err)
			return
		}
		return
	}
}

func DeleteTodo(_ http.ResponseWriter, _ *http.Request) {
}

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

// SetupRouter creates and returns a new HTTP router
func SetupRouter() http.Handler {
	router := &http.ServeMux{}

	// Set up HTTP routes
	router.HandleFunc("GET /", Home)                        // Display homepage
	router.HandleFunc("GET /todo", ReadTodos)               // Return all todo items
	router.HandleFunc("GET /todos", ReadTodos)              // Return all todo items
	router.HandleFunc("GET /todo/{todo_id}", ReadTodo)      // Return a todo item by ID
	router.HandleFunc("POST /todo", CreateTodo)             // Add a todo item and return it
	router.HandleFunc("PUT /todo/{todo_id}", UpdateTodo)    // Change a todo item by ID
	router.HandleFunc("DELETE /todo/{todo_id}", DeleteTodo) // Remove a todo item by ID

	return router
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

	// Create HTTP router
	router := SetupRouter()

	// Get HTTP server port from environment
	port := os.Getenv("TODO_PORT")
	// If the environment variable is empty use port 8080
	if port == "" {
		port = "8080"
	}

	// Create and configure HTTP server
	server := &http.Server{
		Addr:              ":" + port, // Listen on all interfaces on the port defined above
		Handler:           router,
		ReadHeaderTimeout: 2 * time.Second, // Prevent slowloris attack
	}

	// Print a nice message on the terminal
	log.Println("[main] starting server at port 8080")

	// Start the server and in case that fails print an error message
	log.Fatalln("[main] error starting server:", server.ListenAndServe())
}
