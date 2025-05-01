package application

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: Сделать проверку на наличие login
func createTables(ctx context.Context, db *sql.DB) error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id TEXT PRIMARY KEY, 
		jwt TEXT,
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		expression TEXT NOT NULL,
		user_id INTEGER NOT NULL,

		FOREIGN KEY (user_id) REFERENCES expressions (id)
	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		return err
	}

	return nil
}

func AddUser() {
	var q = `
	INSERT INTO expressions (expression, user_id) values ($1, $2)
	`
}

