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

type Request struct {
	Login string `json:"login"`
	Pas   string `json:"pas"`
}

type Rsp struct {
	Status string `json:"status"`
	Jwt    string `json:"jwt"`
}

type JWT struct {
	jwt string
}

// CASE 2
func TestDBC2(t *testing.T) {
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

	var incrs Rsp

	handler := http.HandlerFunc(app.SignUp)
	server := httptest.NewServer(handler)
	defer server.Close()

	reqs := Request{
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
	handler = http.HandlerFunc(app.SignIn)
	server = httptest.NewServer(handler)
	
	incorreq := Request{
		Login: "User",
		Pas:   "sdfghjk",
	}

	body, err = json.Marshal(incorreq)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	req, err = http.NewRequest("POST", server.URL+"/api/v1/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}

	json.NewDecoder(resp.Body).Decode(&incrs)
	
	expectedStatus := "Incorrect password"

	if incrs.Status != expectedStatus {
		t.Fatal("Incorrect login or password")
		return
	}

}
