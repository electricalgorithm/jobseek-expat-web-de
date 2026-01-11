package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./jobseek.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT NOT NULL,
		"email" TEXT NOT NULL UNIQUE,
		"password" TEXT NOT NULL,
		"subscription_plan" TEXT DEFAULT 'basic',
		"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	log.Println("Creating users table...")
	statement, err := DB.Prepare(createTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
	log.Println("Users table created")
}
