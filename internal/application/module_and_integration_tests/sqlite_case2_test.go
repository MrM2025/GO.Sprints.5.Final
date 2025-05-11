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
	Login    string `json:"login"`
	Password string `json:"password"`
}

type NRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Rsp struct {
	Status string `json:"status"`
	Jwt    string `json:"jwt"`
}

type JWT struct {
	Jwt string
}

// CASE 2
func TestDBC2(t *testing.T) {
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

	ap := application.NewOrchestrator(db, ctx)
	ap.CreateTables()

	err = ap.UTD(ctx, "User", db)
	if err != nil {
		t.Fatal(err)
	}

//// SignUp
	handler := http.HandlerFunc(ap.SignUp)
	server := httptest.NewServer(handler)
	defer server.Close()

	reqs := Request{
		Login:    "User",
		Password: "123",
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
	var incrs Rsp

	handl := http.HandlerFunc(ap.SignIn)
	serv := httptest.NewServer(handl)
	defer serv.Close()

	incorreq := NRequest{
		Login:    "User",
		Password: "qwerty",
	}

	bdy, err := json.Marshal(incorreq)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	rq, err := http.NewRequest("POST", serv.URL+"/api/v1/login", bytes.NewBuffer(bdy))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	resp, err := serv.Client().Do(rq)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&incrs)

	expectedStatus := "Incorrect password"

	if incrs.Status != expectedStatus {
		t.Fatal("Incorrect login or password")
		return
	}

}
