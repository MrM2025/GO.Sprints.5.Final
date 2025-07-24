package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/MrM2025/rpforcalc/tree/master/calc_go/internal/application"
	"github.com/MrM2025/rpforcalc/tree/master/calc_go/pkg/errorStore"
	pb "github.com/MrM2025/rpforcalc/tree/master/calc_go/proto"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type fakeAg struct {
	ComputingPower int
	grpcClient     pb.OrchestratorAgentServiceClient
}

func TestIntegratAgent(t *testing.T) {
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

	//// Making a fake gRPC connection
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterOrchestratorAgentServiceServer(grpcSrv, ap)
	go func() {
		grpcSrv.Serve(lis)
	}()

	ap.ExprCounter++
	exprID := strconv.Itoa(ap.ExprCounter)

	ast, err := application.ParseAST("1+1-1+50")
	if err != nil {
		t.Fatal(err)
		return
	}

	expr := &application.Expression{
		ID:     exprID,
		Expr:   "1+1-1+50",
		Jwt:    "qwe.ewq.wqe",
		Login:  "User",
		Status: "pending",
		AST:    ast,
	}

	ap.ExprStore[exprID] = expr
	ap.Tasks(expr)

	a := NewfakeAgent()

	go func() {
		for i := 0; i < a.ComputingPower; i++ {
			if id, err := a.fakeAgent(ap, t); err != nil {
				if !errors.Is(err, errorStore.DvsByZeroErr) {
					log.Fatalf("Task ID: %s, err: %s", id, err)
				}
				log.Printf("Task ID: %s, err: %s", id, err)
			}

			time.Sleep(2 * time.Second)

			if expr = ap.ExprStore[exprID]; expr.Status != "completed" {
				log.Fatalf("The expression - %s hasn't been solved", expr.Expr)
			}
		}
	}()

}

func NewfakeAgent() *fakeAg {
	cp, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil || cp < 1 {
		cp = 1
	}

	grpcAddr := "localhost:9090"

	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewOrchestratorAgentServiceClient(conn)

	return &fakeAg{
		ComputingPower: cp,
		grpcClient:     client,
	}
}

func (a *fakeAg) fakeAgent(ap *application.Orchestrator, t *testing.T) (string, error) {
	var divbyzeroeerr error
	for {
		rs, err := a.grpcClient.Get(ap.Ctx, nil)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		result, diverr := calculator(rs.Operation, rs.Arg1, rs.Arg2)

		if errors.Is(diverr, errorStore.DvsByZeroErr) {
			divbyzeroeerr = errorStore.DvsByZeroErr
		}

		if divbyzeroeerr != nil {
			return rs.Id, divbyzeroeerr
		}

		_, err = a.grpcClient.Post(context.Background(), &pb.PostRequest{Id: rs.Id, Result: result})
		if err != nil {
			log.Fatal(err)
			return rs.Id, err
		}

	}
}

func calculator(operator string, arg1, arg2 float64) (float64, error) {
	var result float64

	switch {
	case operator == "+":
		result = arg1 + arg2
	case operator == "-":
		result = arg1 - arg2
	case operator == "*":
		result = arg1 * arg2
	case operator == "/":
		if arg2 == 0 {
			return 0, errorStore.DvsByZeroErr
		}
		result = arg1 / arg2
	default:
		return 0, fmt.Errorf("Error")
	}

	return result, nil
}
