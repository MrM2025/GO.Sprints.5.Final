package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	//"os"
	//"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Rsp struct {
	Status string `json:"status,omitempty"`
	Jwt    string `json:"jwt,omitempty"`
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Hash struct {
	hash string
}

var LeftIDCounter int
var RightIDCounter int

func (o *Orchestrator) CreateTables() error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		login TEXT UNIQUE NOT NULL,
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT NOT NULL, 
		jwt TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY,
		expression TEXT NOT NULL,
		jwt TEXT NOT NULL,
		user_lg TEXT NOT NULL,
		status TEXT NOT NULL,
		result REAL,
	
		FOREIGN KEY (user_lg) REFERENCES users(login) ON DELETE CASCADE
	);`
	)

	if _, err := o.Db.ExecContext(o.ctx, usersTable); err != nil {
		return err
	}

	if _, err := o.Db.ExecContext(o.ctx, expressionsTable); err != nil {
		return err
	}

	return nil
}

// UTD - Users Table Deleting
func (o *Orchestrator) UTD(ctx context.Context, lg string, db *sql.DB) error {

	_, err := db.ExecContext(ctx, "DELETE FROM users WHERE login=?", lg)
	if err != nil {
		return err
	}

	return nil
}

func hash(p string) (string, error) {
	saltedBytes := []byte(p)
	hashed, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed[:]), nil
}

func compare(hash, p string) error {
	h := []byte(hash)
	ps := []byte(p)
	return bcrypt.CompareHashAndPassword(h, ps)
}

func (o *Orchestrator) AddUser(ctx context.Context, lg, hashed string, db *sql.DB) error {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = db.ExecContext(ctx, "INSERT INTO users(login, hash) VALUES(?, ?)", lg, hashed)

	if err != nil {
		return fmt.Errorf("%s", err)
	}

	return tx.Commit()
}

func (o *Orchestrator) GetExpr(id int, lg string) (*Expression, bool, error) {
	var expr *Expression

	rows, err := o.Db.QueryContext(
		o.ctx,
		`SELECT  expression, jwt, user_lg, status, result FROM expressions WHERE user_lg = $1 AND id = $2`,
		lg,
		id,
	)

	if err != nil {
		return nil, false, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, false, fmt.Errorf("rows error: %w", err)
		}

		return nil, false, nil
	}

	if err = rows.Scan(
		&expr.ID,
		&expr.Expr,
		&expr.Jwt,
		&expr.Login,
		&expr.Status,
		expr.Result,
	); err != nil {
		return nil, false, fmt.Errorf("scan error: %w", err)
	}

	return expr, true, nil

}

func (o *Orchestrator) AddExpr(expr *Expression, rok bool, db *sql.DB) error {

	if !rok {
		q := `INSERT INTO expressions(id, expression, jwt, user_lg, status, result) VALUES(?, ?, ?, ?, ?, ?)`
		_, err := o.Db.ExecContext(o.ctx, q, expr.ID, expr.Expr, expr.Jwt, expr.Login, expr.Status, 1)
		if err != nil {
			return err
		}

		return nil
	}

	up := `UPDATE users SET status = $1, result = $2 WHERE id = $3 AND user_lg = $4`
	_, err := o.Db.ExecContext(o.ctx, up, expr.Status, expr.Result, expr.ID, expr.Login)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) SignIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var (
		u  User
		h  Hash
		id int = 0
	)

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, `{"error":"Invalid Body"}`, http.StatusUnprocessableEntity)
		return
	}

	q := `SELECT hash FROM users WHERE login = ?`
	up := `UPDATE users SET jwt= $1 WHERE login = $2`

	// TODO: если нет логина, сказать, что он неправильный
	err = o.Db.QueryRowContext(o.ctx, q, u.Login).Scan(&h.hash)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("Incorrect login")
			return
		}
	}

	er := compare(h.hash, u.Password)
	if er != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Rsp{Status: "Incorrect password"})
		return
	}

	for {
		id++
		expr, ok, err := o.GetExpr(id, u.Login)
		if !ok {
			if err != nil
		}

		o.exprCounter = id
		expr.ID = strconv.Itoa(id)
		exprStore[expr.ID] = expr

		if expr.Status != "completed" {
			o.Tasks(expr)
		}
	}

	jwt := AddJWT(u.Login)

	for _, expr := range exprStore {
		if expr.Login == u.Login {
			expr.Jwt = jwt
		}
	}

	_, err = o.Db.ExecContext(o.ctx, up, jwt, u.Login)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Rsp{Status: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Rsp{Status: "Successful sign in", Jwt: jwt})
}

func (o *Orchestrator) SignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var u User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, `{"error":"Invalid Body"}`, http.StatusUnprocessableEntity)
		return
	}

	h, err := hash(u.Password)

	if err != nil {
		http.Error(w, fmt.Sprintln(err), http.StatusInternalServerError)
		return
	}

	if err := o.AddUser(o.ctx, u.Login, h, o.Db); err != nil {
		var status int
		switch {
		case errors.Is(err, sql.ErrNoRows):
			status = http.StatusNotFound
		case strings.Contains(err.Error(), "UNIQUE constraint"):
			status = http.StatusConflict
			w.Write([]byte(`{"status":"Login already exists"}`))
		default:
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Rsp{Status: "Successful sign up"})

}

// Delete all tables
func (o *Orchestrator) DTBs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_, err := o.Db.ExecContext(o.ctx, `DELETE FROM users`)
	if err != nil {
		http.Error(w, "deleting all tables error", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Everything has been deleted")
	//json.NewEncoder(w).Encode(`{"status":"Everything has been deleted"}`)

}

/*
func (o *Orchestrator) DDB(w http.ResponseWriter, r *http.Request) {

	err := os.Remove("../teststore.db") or All or ../path/path
	if err != nil {
		http.Error(w, "deleting db error", http.StatusConflict)
		return
	}

	json.NewEncoder(w).Encode("everything has been deleted")

}
*/
