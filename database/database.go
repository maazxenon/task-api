package database

import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the SQLite database and creates the tasks table if it doesn't exist
func InitDB() {
    var err error
    DB, err = sql.Open("sqlite3", "./app.db")
    if err != nil {
        log.Fatal(err)
    }

    sqlStmt := `
    CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        description TEXT,
        due_date TEXT,
        status TEXT
    );`

    _, err = DB.Exec(sqlStmt)
    if err != nil {
        log.Fatalf("Error creating table: %q: %s\n", err, sqlStmt)
    }

    log.Println("Database initialized")
}
