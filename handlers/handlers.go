package handlers

import (
    "database/sql"
    "fmt"
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    _ "github.com/maazxenon/task-api/docs"
    "github.com/maazxenon/task-api/database"
    "log"
    "github.com/maazxenon/task-api/models"
)

// Task represents a task in the task list
type Task models.Task

type ErrorResponse struct {
    Message string `json:"message"`
}

// Validator instance
var validate *validator.Validate

// init initializes the validator
func init() {
    validate = validator.New()
    validate.RegisterValidation("status", validateStatus)
}

// validateStatus is a custom validation function for the status field
func validateStatus(fl validator.FieldLevel) bool {
    status := fl.Field().String()
    switch status {
    case "pending", "completed", "in progress":
        return true
    }
    return false
}

// DB is a global variable for the SQLite database connection
var DB *sql.DB = database.DB

// IndexHandler serves the main page and displays all tasks
// @Summary Get all tasks
// @Description Get a list of all tasks
// @Tags tasks
// @Produce  json
// @Success 200 {array} Task
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /tasks [get]
func IndexHandler(c *gin.Context) {
    rows, err := database.DB.Query("SELECT id, title, description, due_date, status FROM tasks")
    if err != nil {
        log.Printf("Error querying tasks: %v", err)
        c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
        return
    }
    defer rows.Close()

    tasks := []Task{}
    for rows.Next() {
        var task Task
        if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status); err != nil {
            log.Printf("Error scanning task: %v", err)
            c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
            return
        }
        tasks = append(tasks, task)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error with rows: %v", err)
        c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
        return
    }

    c.JSON(http.StatusOK, tasks)
}
// GetTaskHandler handles fetching details of a specific task by ID
// @Summary Get task details
// @Description Get details of a task by ID
// @Tags tasks
// @Produce  json
// @Param id path int true "Task ID"
// @Success 200 {object} Task
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 404 {object} ErrorResponse "Not Found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /tasks/{id} [get]
func GetTaskHandler(c *gin.Context) {
    id := c.Param("id")

    var task Task
    err := database.DB.QueryRow("SELECT id, title, description, due_date, status FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, ErrorResponse{Message: "Task not found"})
        } else {
            log.Printf("Error querying task: %v", err)
            c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, task)
}
// CreateHandler handles the creation of a new task
// @Summary Create a new task
// @Description Create a new task with the provided details
// @Tags tasks
// @Accept  json
// @Produce  json
// @Param task body Task true "Task"
// @Success 200 {object} Task
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /tasks [post]
func CreateHandler(c *gin.Context) {
    var task Task
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
        return
    }

    // Validate the task struct
    if err := validate.Struct(task); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
        return
    }

    result, err := database.DB.Exec("INSERT INTO tasks(title, description, due_date, status) VALUES(?, ?, ?, ?)", task.Title, task.Description, task.DueDate, task.Status)
    if err != nil {
        log.Printf("Error executing query: %v", err)
        c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
        return
    }

    id, err := result.LastInsertId()
    if err != nil {
        log.Printf("Error getting last insert ID: %v", err)
        c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
        return
    }

    task.ID = int(id)
    c.JSON(http.StatusOK, task)
}

// UpdateTaskHandler handles updating an existing task
// @Summary Update a task
// @Description Update the details of an existing task by ID
// @Tags tasks
// @Accept  json
// @Produce  json
// @Param id path int true "Task ID"
// @Param task body Task true "Task"
// @Success 200 {object} Task
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 404 {object} ErrorResponse "Not Found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /tasks/{id} [put]
func UpdateTaskHandler(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid task ID"})
        return
    }

    var task Task
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
        return
    }

    // Validate the task struct
    if err := validate.Struct(task); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
        return
    }

    result, err := database.DB.Exec("UPDATE tasks SET title = ?, description = ?, due_date = ?, status = ? WHERE id = ?", task.Title, task.Description, task.DueDate, task.Status, id)
    if err != nil {
        log.Printf("Error executing query: %v", err)
        c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Printf("Error getting rows affected: %v", err)
        c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
        return
    }

    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, ErrorResponse{Message: "Task not found"})
        return
    }

    task.ID = id
    c.JSON(http.StatusOK, task)
}

// DeleteHandler handles the deletion
// @Summary Delete a task
// @Description Delete a task by ID
// @Tags tasks
// @Produce  json
// @Param id path int true "Task ID"
// @Success 200 {object} map[string]string "message: Task deleted"
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 404 {object} ErrorResponse "Not Found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /tasks/{id} [delete]
func DeleteHandler(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid task ID"})
        return
    }

    err = deleteTaskByID(id)
    if err != nil {
        if err == ErrTaskNotFound {
            c.JSON(http.StatusNotFound, ErrorResponse{Message: "Task not found"})
        } else {
            log.Printf("Error deleting task: %v", err)
            c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal server error"})
        }
        return
    }

    c.JSON(http.StatusOK, map[string]string{"message": "Task deleted"})
}

// deleteTaskByID deletes a task by its ID from the data store
func deleteTaskByID(id int) error {
    result, err := database.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
    if err != nil {
        log.Printf("Error executing delete query: %v", err)
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Printf("Error getting rows affected: %v", err)
        return err
    }

    if rowsAffected == 0 {
        return ErrTaskNotFound
    }

    return nil
}

// ErrTaskNotFound is an error returned when a task is not found
var ErrTaskNotFound = fmt.Errorf("task not found")