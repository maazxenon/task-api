package models


var TaskTable = `
 CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        description TEXT,
        due_date TEXT,
        status TEXT
    );`



type Task struct {
		ID          int    `json:"id" example:"1"`
		Title       string `json:"title" example:"Buy groceries" binding:"required"`
		Description string `json:"description" example:"Milk, Bread, Cheese"`
		DueDate     string `json:"due_date" example:"2023-12-31"`
		Status      string `json:"status" example:"pending" validate:"required,status"`
}
	










