package main

import (
    "log"
    "github.com/maazxenon/task-api/routes"
    "github.com/maazxenon/task-api/database"
)

func main() {
    // Initialize the database
    database.InitDB()
    defer database.DB.Close()

    // Set up the router
    r := routes.TaskRouter()
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}