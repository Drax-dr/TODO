// main.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
  	"github.com/gin-contrib/cors"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"time"
)

var ldb *leveldb.DB

type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

func setupRouter() *gin.Engine {
	r := gin.Default()
  r.Use(cors.Default())
	r.GET("/todos", getTodos)
	r.POST("/todos", createTodo)
	r.GET("/todos/:id", getTodo)
	r.PUT("/todos/:id", updateTodo)
	r.DELETE("/todos/:id", deleteTodo)

	return r
}

func getTodos(c *gin.Context) {
	todos := getAllTodos()

	c.JSON(200, todos)
}

func createTodo(c *gin.Context) {
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	todo.ID = generateID()
	todo.CreatedAt = time.Now()

	saveTodo(todo)

	c.JSON(200, todo)
}

func getTodo(c *gin.Context) {
	id := c.Params.ByName("id")
	todo, err := getTodoByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Todo not found"})
		return
	}
	c.JSON(200, todo)
}

func updateTodo(c *gin.Context) {
	id := c.Params.ByName("id")
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	existingTodo, err := getTodoByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Todo not found"})
		return
	}

	todo.ID = existingTodo.ID
	todo.CreatedAt = existingTodo.CreatedAt

	saveTodo(todo)

	c.JSON(200, todo)
}

func deleteTodo(c *gin.Context) {
	id := c.Params.ByName("id")
	if err := deleteTodoByID(id); err != nil {
		c.JSON(404, gin.H{"error": "Todo not found"})
		return
	}
	c.JSON(200, gin.H{"id #" + id: "deleted"})
}

func getAllTodos() []Todo {
	var todos []Todo

	iter := ldb.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		var todo Todo
		if err := json.Unmarshal(iter.Value(), &todo); err == nil {
			todos = append(todos, todo)
		}
	}

	return todos
}

func saveTodo(todo Todo) {
	value, _ := json.Marshal(todo)
	ldb.Put([]byte(todo.ID), value, nil)
}

func getTodoByID(id string) (Todo, error) {
	var todo Todo

	data, err := ldb.Get([]byte(id), nil)
	if err != nil {
		return todo, err
	}

	if err := json.Unmarshal(data, &todo); err != nil {
		return todo, err
	}

	return todo, nil
}

func deleteTodoByID(id string) error {
	return ldb.Delete([]byte(id), nil)
}

func generateID() string {
	return time.Now().Format("20060102150405")
}

func main() {
	var err error
	ldb, err = leveldb.OpenFile("leveldb-data", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening LevelDB: %v\n", err)
		os.Exit(1)
	}
	defer ldb.Close()

	r := setupRouter()
	r.Run(":8080")
}
