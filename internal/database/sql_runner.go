package database

import (
	"database/sql"
	"fmt"
	"os"
)

func RunSQLFile(db *sql.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("execute %s: %w", path, err)
	}
	return nil
}
