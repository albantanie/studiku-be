package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"studi-ku-backend/config"
)

var DB *sql.DB

func Connect(cfg *config.Config) {
	var err error
	DB, err = sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to PostgreSQL database")
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
