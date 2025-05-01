package application

import (
	"context"
	"database/sql"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Login string `json:"login"`
	Pas  string `json:"pas"`
}										

// TODO: Сделать проверку на наличие login
func createTables(ctx context.Context, db *sql.DB) error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		login TEXT PRIMARY KEY, 
		id TEXT
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

func (u *User) AddUser(id, jwt string, db *sql.DB) {
	var q = `
	INSERT INTO users (expression, user_id) values ($1, $2)
	`
	result, err := db.Exec(q, u.Login, u.Pas)
	if err != nil {
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func SighIn(w http.ResponseWriter, r *http.Request) {

}

func (u *User) SignUp(w http.ResponseWriter, r *http.Request) {
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil{
		http.Error(w, `{"error":"Invalid Body"}`, http.StatusUnprocessableEntity)
		return
	}

}