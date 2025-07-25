package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/MrM2025/rpforcalc/tree/master/calc_go/pkg/errorStore"
	pb "github.com/MrM2025/rpforcalc/tree/master/calc_go/proto"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
)

type Config struct {
	Addr                string
	TimeAddition        int
	TimeSubtraction     int
	TimeMultiplications int
	TimeDivisions       int
}

func ConfigFromEnv() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ta, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	if ta == 0 {
		ta = 100
	}
	ts, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	if ts == 0 {
		ts = 100
	}
	tm, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	if tm == 0 {
		tm = 1000
	}
	td, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
	if td == 0 {
		td = 1000
	}

	return &Config{
		Addr:                port,
		TimeAddition:        ta,
		TimeSubtraction:     ts,
		TimeMultiplications: tm,
		TimeDivisions:       td,
	}
}

type Orchestrator struct {
	pb.UnsafeOrchestratorAgentServiceServer
	Config      *Config
	Db          *sql.DB
	ExprStore   map[string]*Expression
	Ctx         context.Context
	taskStore   map[string]*Task
	taskQueue   []*Task
	mu          sync.Mutex
	ExprCounter int
	taskCounter int
}

func NewOrchestrator(db *sql.DB, ctx context.Context) *Orchestrator {
	return &Orchestrator{
		Config:      ConfigFromEnv(),
		Db:          db,
		Ctx:         ctx,
		ExprStore:   make(map[string]*Expression),
		ExprCounter: 0,
		taskStore:   make(map[string]*Task),
		taskQueue:   make([]*Task, 0),
	}
}

type OrchReqJSON struct {
	Expression string `json:"expression"`
	Login      string `json:"login,omitempty"`
	JWT        string `json:"jwt,omitempty"`
}

type OrchResJSON struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type Expression struct {
	ID     string   `json:"id,omitempty"`
	Expr   string   `json:"expression,omitempty"`
	Jwt    string   `json:"-"`
	Login  string   `json:"login,omitempty"`
	Status string   `json:"status,omitempty"`
	Result string   `json:"result,omitempty"`
	AST    *ASTNode `json:"-"`
}

type Task struct {
	ID             string   `json:"id,omitempty"`
	ExprID         string   `json:"expression,omitempty"`
	Arg1           float64  `json:"arg1,omitempty"`
	Arg2           float64  `json:"arg2,omitempty"`
	Operation      string   `json:"operation,omitempty"`
	Operation_time int      `json:"operation_time,omitempty"`
	Node           *ASTNode `json:"-"`
}

var (
	calc TCalc
)

func (o *Orchestrator) Tasks(expr *Expression) {
	var traverse func(node *ASTNode)
	traverse = func(node *ASTNode) {

		if node == nil || node.IsLeaf {
			return
		}

		traverse(node.Left)
		traverse(node.Right)
		if node.Left != nil && node.Right != nil && node.Left.IsLeaf && node.Right.IsLeaf {
			if !node.TaskScheduled {
				o.taskCounter++
				taskID := strconv.Itoa(o.taskCounter)
				var opTime int
				switch node.Operator {
				case "+":
					opTime = o.Config.TimeAddition
				case "-":
					opTime = o.Config.TimeSubtraction
				case "*":
					opTime = o.Config.TimeMultiplications
				case "/":
					opTime = o.Config.TimeDivisions
				default:
					opTime = 100
				}

				task := &Task{
					ID:             taskID,
					ExprID:         expr.ID,
					Arg1:           node.Left.Value,
					Arg2:           node.Right.Value,
					Operation:      node.Operator,
					Operation_time: opTime,
					Node:           node,
				}
				node.TaskScheduled = true
				o.taskStore[taskID] = task
				o.taskQueue = append(o.taskQueue, task)
			}
		}
	}
	traverse(expr.AST)
}

func (o *Orchestrator) CalcHandler(w http.ResponseWriter, r *http.Request) { //Сервер, который принимает арифметическое выражение, переводит его в набор последовательных задач и обеспечивает порядок их выполнения.
	var (
		emsg string
		jwt  string
	)
	o.mu.Lock()
	defer o.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	request := new(OrchReqJSON)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body) //Достаем выражение
	dec.DisallowUnknownFields()
	err := dec.Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = o.Db.QueryRowContext(o.Ctx, "SELECT jwt FROM users WHERE login = ?", request.Login).Scan(&jwt); jwt == "" {
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("Session time is up, please, sign in again")
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Incorrect login")
		return
	}

	err = strimJWT(request.Login, request.JWT)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Session time is up, please, sign in again")
		return
	} else if err == nil && request.JWT != jwt {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Incorrect jwt(probably from other user)")
		return
	}

	ok, err := calc.IsCorrectExpression(request.Expression) // Проверяем выражение на наличие ошибок

	if !ok && err != nil { // Присваиваем ошибкам статус-код, выводим их
		switch {
		case errors.Is(err, errorStore.EmptyExpressionErr):
			emsg = errorStore.EmptyExpressionErr.Error()

		case errors.Is(err, errorStore.IncorrectExpressionErr):
			emsg = errorStore.IncorrectExpressionErr.Error()

		case errors.Is(err, errorStore.NumToPopMErr): // numtopop > nums' slise length
			emsg = errorStore.NumToPopMErr.Error()

		case errors.Is(err, errorStore.NumToPopZeroErr): // numtopop <= 0
			emsg = errorStore.NumToPopZeroErr.Error()

		case errors.Is(err, errorStore.NthToPopErr): // no operator to pop
			emsg = errorStore.NthToPopErr.Error()

		case errors.Is(err, errorStore.DvsByZeroErr):
			emsg = errorStore.DvsByZeroErr.Error()
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(OrchResJSON{Error: emsg})
		return
	}

	o.ExprCounter++
	exprID := strconv.Itoa(o.ExprCounter)

	ast, err := ParseAST(request.Expression)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnprocessableEntity)
		return
	}

	expr := &Expression{
		ID:     exprID,
		Expr:   request.Expression,
		Jwt:    request.JWT,
		Login:  request.Login,
		Status: "pending",
		AST:    ast,
	}

	o.ExprStore[exprID] = expr
	o.Tasks(expr)

	err = o.AddExpr(expr, false, o.Db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(`Sorry, something went wrong, try again later`)
		log.Fatal(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(OrchResJSON{ID: exprID})

}

func (o *Orchestrator) Get(ctx context.Context, _ *pb.Empty) (*pb.GetResponse, error) {

	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.taskQueue) == 0 {
		return &pb.GetResponse{}, fmt.Errorf("No task available")
	}

	task := o.taskQueue[0]
	o.taskQueue = o.taskQueue[1:]

	if expr, exists := o.ExprStore[task.ExprID]; exists {
		expr.Status = "in_progress"
	}

	return &pb.GetResponse{Id: task.ID, Arg1: task.Arg1, Arg2: task.Arg2, Operation: task.Operation, OperationTime: int32(task.Operation_time)}, nil
}

func (o *Orchestrator) Post(ctx context.Context, in *pb.PostRequest) (*pb.Empty, error) {

	o.mu.Lock()
	task, ok := o.taskStore[in.Id]

	if !ok {
		o.mu.Unlock()
		return nil, fmt.Errorf("No task available")
	}

	task.Node.IsLeaf = true
	task.Node.Value = in.Result
	delete(o.taskStore, in.Id)

	if expr, exists := o.ExprStore[task.ExprID]; exists {
		o.Tasks(expr)
		if expr.AST.IsLeaf {
			expr.Status = "completed"
			expr.Result = strconv.FormatFloat(expr.AST.Value, 'g', 8, 32)
		}

		err := o.AddExpr(expr, true, o.Db)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

	}

	o.mu.Unlock()

	return nil, nil
}

func makeAnAtomicExpr(Operation string, Arg1, Arg2 float64) (string, error) {
	arg1 := strconv.FormatFloat(Arg1, 'g', 8, 32)
	arg2 := strconv.FormatFloat(Arg2, 'g', 8, 32)

	var result string

	switch {
	case Operation == "+":
		result = arg1 + "+" + arg2
	case Operation == "-":
		result = arg1 + "-" + arg2
	case Operation == "*":
		result = arg1 + "*" + arg2
	case Operation == "/":
		if arg2 == "0" {
			return "0", errorStore.DvsByZeroErr //DvsByZeroErr
		}
		result = arg1 + "/" + arg2
	}
	return result, nil
}

func (o *Orchestrator) RunOrchestrator() {
	a := NewAgent()

	go func() {
		for i := 0; i < a.ComputingPower; i++ {
			a.worker()
		}
	}()

	mux := http.NewServeMux()
	http.Handle("/", mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { //можно открыть README.md
		http.ServeFile(w, r, "..\\README.md")
	})
	mux.HandleFunc("/api/v1/calculate", o.CalcHandler)
	mux.HandleFunc("/api/v1/expressions", o.ExpressionsOutput)
	mux.HandleFunc("/api/v1/expression/id", o.ExpressionByID)
	mux.HandleFunc("/api/v1/register", o.SignUp)
	mux.HandleFunc("/api/v1/login", o.SignIn)
	mux.HandleFunc("/api/v1/DTBs", o.DTBs)
	//mux.HandleFunc("/api/v1/DDB", o.DDB)

	go func() {
		for {
			time.Sleep(2 * time.Second)
			o.mu.Lock()
			if len(o.taskQueue) > 0 {
				log.Printf("Pending tasks in queue: %d", len(o.taskQueue))
			}
			o.mu.Unlock()
		}
	}()

	go func() {
		log.Println("HTTP listening on", o.Config.Addr)
		if err := http.ListenAndServe(":"+o.Config.Addr, mux); err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterOrchestratorAgentServiceServer(grpcSrv, o)
	log.Println("gRPC listening on 9090")
	grpcSrv.Serve(lis)
}
