package application

import (
	"bytes"
	"context"

	//"log"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrM2025/rpforcalc/tree/master/calc_go/internal/application"
	_ "github.com/mattn/go-sqlite3"
)

type Reqs struct {
	Login string `json:"login"`
	Pas   string `json:"pas"`
}

type Resp struct {
	Status string `json:"status"`
	Jwt    string `json:"jwt"`
}

type JSONWebToken struct {
	Jwt string
}

// CASE 1
func TestDBC1(t *testing.T) {
	//// Deleting the db tables for a new test
	ctx := context.TODO()

	db, err := sql.Open("sqlite3", "teststore.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		panic(err)
	}

	app := application.NewOrchestrator(db, ctx)
	app.CreateTables()

	err = app.UTD(ctx, "User", db)
	if err != nil {
		t.Fatal(err)
	}

	//// SignUp
	handler := http.HandlerFunc(app.SignUp)
	server := httptest.NewServer(handler)
	defer server.Close()

	reqs := Reqs{
		Login: "User",
		Pas:   "123",
	}

	body, err := json.Marshal(reqs)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	req, err := http.NewRequest("POST", server.URL+"/api/v1/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	res, err := server.Client().Do(req)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 , but got %d", res.StatusCode)
	}

	//// SignIn
	var (
		rs Resp
		j  JSONWebToken
	)

	handl := http.HandlerFunc(app.SignIn)
	serv := httptest.NewServer(handl)
	defer serv.Close()

	req, err = http.NewRequest("POST", serv.URL+"/api/v1/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	res, err = serv.Client().Do(req)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&rs)

	qj := `SELECT jwt FROM users WHERE login = ?`
	db.QueryRowContext(ctx, qj, reqs.Login).Scan(&j.Jwt)

	expectedStatus := "Successful sign in"

	if rs.Status != expectedStatus {
		t.Fatal("Incorrect login or password")
		return
	}

	if rs.Jwt != j.Jwt {
		t.Fatal("Incorrect JSON Web Token")
		return
	}

}
