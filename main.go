package main

import (
    "encoding/json"
    "fmt"
    "strconv"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/go-playground/validator/v10"
    "github.com/swaggo/http-swagger"
    _ "task-api/docs"
    "task-api/database"
    "database/sql"
    "log"

)

// Tasl represents a task in the task list
type Task struct {
    ID          int    `json:"id" example:"1"`
    Title       string `json:"title" example:"Buy groceries" binding:"required"`
    Description string `json:"description" example:"Milk, Bread, Cheese"`
    DueDate     string `json:"due_date" example:"2023-12-31"`
    Status      string `json:"status" example:"pending" validate:"required,status"`
}

// ErrorResponse represents an error response
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


// indexHandler serves the main page and displays all tasks
// @Summary Get all tasks
// @Description Get a list of all tasks
// @Tags tasks
// @Produce  json
// @Success 200 {array} Task
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /tasks [get]
func indexHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, title, description, due_date, status FROM tasks")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    tasks := []Task{}
    for rows.Next() {
        var task Task
        if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        tasks = append(tasks, task)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}



// getTaskHandler handles fetching details of a specific task by ID
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
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var task Task
    err := db.QueryRow("SELECT id, title, description, due_date, status FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Task not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}

// createHandler handles the creation of a new task
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
func createHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        var task Task
        err := json.NewDecoder(r.Body).Decode(&task)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Validate the task struct
        err = validate.Struct(task)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        result, err := db.Exec("INSERT INTO tasks(title, description, due_date, status) VALUES(?, ?, ?, ?)", task.Title, task.Description, task.DueDate, task.Status)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        id, err := result.LastInsertId()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        task.ID = int(id)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(task)
    }
}


// updateTaskHandler handles updating an existing task
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
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]

    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    var task Task
    err = json.NewDecoder(r.Body).Decode(&task)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Validate the task struct
    err = validate.Struct(task)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result, err := db.Exec("UPDATE tasks SET title = ?, description = ?, due_date = ?, status = ? WHERE id = ?", task.Title, task.Description, task.DueDate, task.Status, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if rowsAffected == 0 {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }

    task.ID = id

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}

// deleteHandler handles the deletion of a task
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
func deleteHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "DELETE" {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    err = deleteTaskByID(id)
    if err != nil {
        if err == ErrTaskNotFound {
            http.Error(w, "Task not found", http.StatusNotFound)
        } else {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted"})
}

// deleteTaskByID deletes a task by its ID from the data store
func deleteTaskByID(id int) error {
    // Implement the logic to delete the task from the data store
    // Return ErrTaskNotFound if the task does not exist
    // Return nil if the task is successfully deleted
    // Return other errors if there is an issue with the deletion process

    result, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return ErrTaskNotFound
    }

    return nil

}


// ErrTaskNotFound is an error returned when a task is not found
var ErrTaskNotFound = fmt.Errorf("task not found")





func main() {
    db := database.InitDB()
    defer db.Close()

    r := mux.NewRouter()

    // Serve Swagger UI
    r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

    r.HandleFunc("/tasks", indexHandler).Methods("GET")
    r.HandleFunc("/tasks", createHandler).Methods("POST")
    r.HandleFunc("/tasks/{id}", getTaskHandler).Methods("GET")
    r.HandleFunc("/tasks/{id}", updateTaskHandler).Methods("PUT")
    r.HandleFunc("/tasks/{id}", deleteHandler).Methods("DELETE")

    fmt.Println("Server is running at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}