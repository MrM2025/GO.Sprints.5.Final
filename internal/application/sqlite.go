package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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
		user_id INTEGER NOT NULL,
	
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`
	)

	if _, err := o.Db.ExecContext(o.Ctx, usersTable); err != nil {
		return err
	}

	if _, err := o.Db.ExecContext(o.Ctx, expressionsTable); err != nil {
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

func (o *Orchestrator) AddExpr(expr *Expression, rok bool, db *sql.DB) error {
	var ID int
	id, err := strconv.Atoi(expr.ID)
	if err != nil {
		log.Fatal(err)
	}

	sl := `SELECT id FROM users WHERE login = ?`

	o.Db.QueryRowContext(o.Ctx, sl, expr.Login).Scan(&ID)

	up := `UPDATE expressions SET expression = $1, jwt = $2, user_lg = $3, status = $4, user_id = $5 WHERE id = $6`

	if !rok {
		q := `INSERT INTO expressions(id, expression, jwt, user_lg, status, user_id) VALUES(?, ?, ?, ?, ?, ?)`
		_, err := o.Db.ExecContext(o.Ctx, q, id, expr.Expr, expr.Jwt, expr.Login, expr.Status, ID)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: expressions.id") {
				_, err := o.Db.ExecContext(o.Ctx, up, expr.Expr, expr.Jwt, expr.Login, expr.Status, ID, expr.ID)
				return err
			}
			return err
		}

		return nil
	}

	up = `UPDATE expressions SET status = $1, result = $2 WHERE id = $3 AND user_lg = $4`
	_, err = o.Db.ExecContext(o.Ctx, up, expr.Status, expr.Result, id, expr.Login)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) SignIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var (
		u User
		h Hash
	)

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, `{"error":"Invalid Body"}`, http.StatusUnprocessableEntity)
		return
	}

	q := `SELECT hash FROM users WHERE login = ?`
	up := `UPDATE users SET jwt= $1 WHERE login = $2`

	err = o.Db.QueryRowContext(o.Ctx, q, u.Login).Scan(&h.hash)
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

	jwt := AddJWT(u.Login)

	for _, expr := range o.ExprStore {
		if expr.Login == u.Login {
			expr.Jwt = jwt
		}
	}

	_, err = o.Db.ExecContext(o.Ctx, up, jwt, u.Login)
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

	if err := o.AddUser(o.Ctx, u.Login, h, o.Db); err != nil {
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

	_, err := o.Db.ExecContext(o.Ctx, `DELETE FROM users`)
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
