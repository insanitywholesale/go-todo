# go-todo
Demo REST API in Go written for an IEEE workshop

## Running locally
First [install Go](https://go.dev/doc/install), then [clone the project repository](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository) and finally run it using:
```
go run main.go
```

The database file, todo.db, will be created automatically if it doesn't exist already.

## `curl` commands
You can use [cURL](https://curl.se/) to make HTTP requests from the command line and interact with the REST API.

### GET all
To get all TODO items:
```
curl -s -X GET 'http://localhost:8080/todos' | jq
```

### GET one
To get all TODO items by its ID (for example one with ID 2):
```
curl -s -X GET 'http://localhost:8080/todo/2' | jq
```

### POST
To create some TODO items:
```
curl -s -X POST -d '{"description":"number 1 test todo", "done":false}' 'http://localhost:8080/todo' | jq
curl -s -X POST -d '{"description":"number 2 test todo", "done":false}' 'http://localhost:8080/todo' | jq
curl -s -X POST -d '{"description":"number 3 test todo", "done":false}' 'http://localhost:8080/todo' | jq
curl -s -X POST -d '{"description":"number 4 test todo", "done":false}' 'http://localhost:8080/todo' | jq
curl -s -X POST -d '{"description":"number 5 test todo", "done":false}' 'http://localhost:8080/todo' | jq
curl -s -X POST -d '{"description":"number 60 test todo", "done":true}' 'http://localhost:8080/todo' | jq
curl -s -X POST -d '{"description":"number 7 test todo", "done":true}'  'http://localhost:8080/todo' | jq
```

### PUT
To change a TODO item by its ID (for example the one with ID 6):
```
curl -s -X PUT -d '{"description":"number 6 test todo","done":true}' 'http://localhost:8080/todo/6' | jq
```

### DELETE
To DELETE a TODO item by its ID (for example the one with ID 7):
```
curl -s -X DELETE 'http://localhost:8080/todo/7' | jq
```
