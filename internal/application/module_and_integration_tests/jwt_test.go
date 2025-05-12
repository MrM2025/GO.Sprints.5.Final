package application

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	//"log"
	"testing"

	"github.com/MrM2025/rpforcalc/tree/master/calc_go/internal/application"
	_ "github.com/mattn/go-sqlite3"
)

type OrchReqJSON struct {
	Expression string `json:"expression"`
	Login      string `json:"login,omitempty"`
	JWT        string `json:"jwt,omitempty"`
}

type IDForExpression struct {
	ID  string `json:"id,omitempty"`
	JWT string `json:"jwt,omitempty"`
}

type User1 struct {
	Expr     string `json:"expression,omitempty"`
	ID       int    `json:"id,omitempty"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password"`
	JWT      string `json:"jwt,omitempty"`
}

type User2 struct {
	Expr     string `json:"expression,omitempty"`
	ID       int    `json:"id,omitempty"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password"`
	JWT      string `json:"jwt,omitempty"`
}

type IDRps struct {
	ID     string `json:"id,omitempty"`
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
}

type ExprResp struct {
	Expression application.Expression `json:"expression,omitempty"`
}

type SlExprsResp struct {
	Expressions []application.Expression `json:"expression,omitempty"`
}

func TestWithTwoUsers1(t *testing.T) {
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

	err = app.UTD(ctx, "User1", db)
	if err != nil {
		t.Fatal(err)
	}

	err = app.UTD(ctx, "User2", db)
	if err != nil {
		t.Fatal(err)
	}

	//// SignUp
	handler := http.HandlerFunc(app.SignUp)
	server := httptest.NewServer(handler)
	defer server.Close()

	reqt1 := User1{
		Login:    "User1",
		Password: "123",
	}

	reqt2 := User2{
		Login:    "User2",
		Password: "321",
	}

	body1, err := json.Marshal(reqt1)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	body2, err := json.Marshal(reqt2)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	req1, err := http.NewRequest("POST", server.URL+"/api/v1/register", bytes.NewBuffer(body1))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	req2, err := http.NewRequest("POST", server.URL+"/api/v1/register", bytes.NewBuffer(body2))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	res1, err := server.Client().Do(req1)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res1.Body.Close()

	res2, err := server.Client().Do(req2)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res2.Body.Close()

	if res1.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 , but got %d", res1.StatusCode)
	}

	if res1.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 , but got %d", res2.StatusCode)
	}

	//// SignIn
	var (
		rs1 Resp
		rs2 Resp
		j1  JSONWebToken
		j2  JSONWebToken
	)

	handl := http.HandlerFunc(app.SignIn)
	serv := httptest.NewServer(handl)
	defer serv.Close()

	req1, err = http.NewRequest("POST", serv.URL+"/api/v1/login", bytes.NewBuffer(body1))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	req2, err = http.NewRequest("POST", serv.URL+"/api/v1/login", bytes.NewBuffer(body2))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	res1, err = serv.Client().Do(req1)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res1.Body.Close()

	res2, err = serv.Client().Do(req2)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res2.Body.Close()

	err = json.NewDecoder(res1.Body).Decode(&rs1)
	if err != nil {
		t.Fatal(err)
	}

	err = json.NewDecoder(res2.Body).Decode(&rs2)
	if err != nil {
		t.Fatal(err)
	}

	qj := `SELECT jwt FROM users WHERE login = ?`
	err = db.QueryRowContext(ctx, qj, reqt1.Login).Scan(&j1.Jwt)
	if err != nil {
		t.Fatal(err)
	}

	err = db.QueryRowContext(ctx, qj, reqt2.Login).Scan(&j2.Jwt)
	if err != nil {
		t.Fatal(err)
	}

	expectedStatus := "Successful sign in"

	if rs1.Status != expectedStatus {
		t.Fatal("Incorrect login or password")
		return
	}

	if rs2.Status != expectedStatus {
		t.Fatal("Incorrect login or password")
		return
	}

	if rs1.Jwt != j1.Jwt {
		t.Fatal("Incorrect JSON Web Token")
		return
	}

	if rs2.Jwt != j2.Jwt {
		t.Fatal("Incorrect JSON Web Token")
		return
	}

	//// Adding the expression with specific user
	var (
		rsp1 IDRps
		rsp2 IDRps
	)

	handler = http.HandlerFunc(app.CalcHandler)
	server = httptest.NewServer(handler)
	defer server.Close()

	rq1 := OrchReqJSON{
		Expression: "2*2",
		Login:      reqt1.Login,
		JWT:        rs1.Jwt,
	}

	rq2 := OrchReqJSON{
		Expression: "2/2",
		Login:      reqt2.Login,
		JWT:        rs2.Jwt,
	}

	body1, err = json.Marshal(rq1)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	body2, err = json.Marshal(rq2)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	req1, err = http.NewRequest("POST", server.URL+"/api/v1/calculate", bytes.NewBuffer(body1))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	req2, err = http.NewRequest("POST", server.URL+"/api/v1/calculate", bytes.NewBuffer(body2))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	res1, err = server.Client().Do(req1)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res1.Body.Close()

	res2, err = server.Client().Do(req2)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res2.Body.Close()

	err = json.NewDecoder(res1.Body).Decode(&rsp1)
	if err != nil {
		t.Fatal(err)
	}

	err = json.NewDecoder(res2.Body).Decode(&rsp2)
	if err != nil {
		t.Fatal(err)
	}

	if rsp1.Status != "" || rsp1.Error != "" {
		t.Fatal(rsp1.Status)
	}

	if rsp1.Status != "" || rsp1.Error != "" {
		t.Fatal(rsp2.Status)
	}

	//// Comparing all the Users' expressions
	var (
		rp1 ExprResp
		rp2 ExprResp
	)
	handler = http.HandlerFunc(app.ExpressionByID)
	server = httptest.NewServer(handler)
	defer server.Close()

	rqs1 := IDForExpression{
		ID:  rsp1.ID,
		JWT: rs1.Jwt,
	}

	rqs2 := IDForExpression{
		ID:  rsp2.ID,
		JWT: rs2.Jwt,
	}

	body1, err = json.Marshal(rqs1)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	body2, err = json.Marshal(rqs2)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	req1, err = http.NewRequest("POST", server.URL+"/api/v1/calculate", bytes.NewBuffer(body1))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	req2, err = http.NewRequest("POST", server.URL+"/api/v1/calculate", bytes.NewBuffer(body2))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	res1, err = server.Client().Do(req1)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res1.Body.Close()

	res2, err = server.Client().Do(req2)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res2.Body.Close()

	err = json.NewDecoder(res1.Body).Decode(&rp1)
	if err != nil {
		t.Fatal(err)
	}

	err = json.NewDecoder(res2.Body).Decode(&rp2)
	if err != nil {
		t.Fatal(err)
	}

	if rp1.Expression.Expr != "2*2" {
		t.Fatal("Incorrect expression")
	}

	if rp2.Expression.Expr != "2/2" {
		t.Fatal("Incorrect expression")
	}

	//// Posting all the Users' expressions
	var (
		rps1 SlExprsResp
		rps2 SlExprsResp
	)
	handl = http.HandlerFunc(app.ExpressionsOutput)
	serv = httptest.NewServer(handl)
	defer server.Close()

	rqst1 := IDForExpression{
		ID:  rsp1.ID,
		JWT: rs1.Jwt,
	}

	rqst2 := IDForExpression{
		ID:  rsp2.ID,
		JWT: rs2.Jwt,
	}

	body1, err = json.Marshal(rqst1)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	body2, err = json.Marshal(rqst2)
	if err != nil {
		t.Fatal("Marshalling error")
		return
	}

	req1, err = http.NewRequest("POST", serv.URL+"/api/v1/calculate", bytes.NewBuffer(body1))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	req2, err = http.NewRequest("POST", serv.URL+"/api/v1/calculate", bytes.NewBuffer(body2))
	if err != nil {
		t.Fatal("Creating a request error")
		return
	}

	res1, err = server.Client().Do(req1)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res1.Body.Close()

	res2, err = server.Client().Do(req2)
	if err != nil {
		t.Fatal("Request processing error:", err)
		return
	}
	defer res2.Body.Close()

	err = json.NewDecoder(res1.Body).Decode(&rps1)
	if err != nil {
		t.Fatal(err)
	}

	err = json.NewDecoder(res2.Body).Decode(&rps2)
	if err != nil {
		t.Fatal(err)
	}

	if rps1.Expressions[0].Expr != "2*2" {
		t.Fatal("Incorrect expression")
	}

	if rps2.Expressions[0].Expr != "2/2" {
		t.Fatal("Incorrect expression")
	}

}
