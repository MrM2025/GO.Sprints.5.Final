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

type Resp struct {
	status string
}

func TestDB(t *testing.T) {

	ctx := context.TODO()

	db, err := sql.Open("sqlite3", "store.db")
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

	handler := http.HandlerFunc(app.SignUp)
	server := httptest.NewServer(handler)
	defer server.Close()

	type Request struct {
		Login string `json:"login"`
		Pas   string `json:"pas"`
	}

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

	handl := http.HandlerFunc(app.SignIn)
	serv := httptest.NewServer(handl)
	
	


}
