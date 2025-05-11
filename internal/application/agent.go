package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/MrM2025/rpforcalc/tree/master/calc_go/pkg/errorStore"
	pb "github.com/MrM2025/rpforcalc/tree/master/calc_go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
)

type AgentTask struct {
	ID             string  `json:"id,omitempty"`
	ExprID         string  `json:"expression,omitempty"`
	Arg1           float64 `json:"arg1,omitempty"`
	Arg2           float64 `json:"arg2,omitempty"`
	Operation      string  `json:"operation,omitempty"`
	Operation_time int     `json:"operation_time,omitempty"`
	Result         string  `json:"result,omitempty"`
}

type AgentResJSON struct {
	ID     string  `json:"ID,omitempty"`
	Result float64 `json:"result,omitempty"`
}

type Agent struct {
	ComputingPower int
	grpcClient     pb.OrchestratorAgentServiceClient
}

func NewAgent() *Agent {
	cp, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil || cp < 1 {
		cp = 1
	}

	grpcAddr := "localhost:8080"

	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewOrchestratorAgentServiceClient(conn)	

	return &Agent{
		ComputingPower: cp,
		grpcClient:     client,
	}
}

func (a *Agent) worker() {
	for {
		task, err := a.grpcClient.Get(context.Background(), &pb.Empty{})
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		log.Printf("Worker: received task %s: %f %s %f, simulating %d ms", task.Id, task.Arg1, task.Operation, task.Arg2, task.OperationTime)
		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
		divbyzeroeerr = nil
		result, diverr := calculator(task.Operation, task.Arg1, task.Arg2)

		if errors.Is(diverr, errorStore.DvsByZeroErr) {
			divbyzeroeerr = errorStore.DvsByZeroErr
		}

		_, err = a.grpcClient.Post(context.Background(), &pb.PostRequest{Id: task.Id, Result: result})
		if err != nil {
			log.Fatal(err)
			return
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

/*
func (a *Agent) SendResult(request *AgentTask, result []byte) {

	_, err := strconv.Atoi(request.ID)
	if err != nil {
		log.Printf("Error of type conversion %v", err)
	}

	req, err := http.NewRequest("POST", a.OrchestratorURL+"/internal/task", bytes.NewReader(result))
	if err != nil {
		log.Printf("Error fetching task: %v. Retrying in 2 seconds...", err)
		time.Sleep(2 * time.Second)

	}

	_, err = a.client.Do(req)
	if err != nil {
		log.Printf("Error doing request: %v. Retrying in 2 seconds...", err)

	}


		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			log.Printf("Worker : error response posting result for task %v: %s", taskStore[ID-1], string(body))
		} else {
			log.Printf("Worker : successfully completed task %v with result %s", taskStore[ID-1], result)
		}

	defer req.Body.Close()

}
*/

func (a *Agent) RunAgent() {
	for i := 0; i < a.ComputingPower; i++ {
		log.Printf("Starting worker %d", i)
		go a.worker()
	}
	select {}
}
