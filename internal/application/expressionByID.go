package application

import (
	"encoding/json"
	"math"
	"net/http"
	"sync"
	//"strconv"
)

type ExprResp struct {
	Expression *Expression `json:"expression,omitempty"`
}

type IDForExpression struct {
	ID  string `json:"id,omitempty"`
	JWT string `json:"jwt,omitempty"`
}

var (
	mu sync.Mutex
)

func (o *Orchestrator) ExpressionByID(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	request := new(IDForExpression)
	json.NewDecoder(r.Body).Decode(&request)

	err := strimJWT("User", request.JWT)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Session time is up, please, sign in again")
		return
	}

	expr, ok := o.ExprStore[request.ID]

	if !ok || request.JWT != expr.Jwt {
		http.Error(w, `{"error":"Expression not found"}`, http.StatusNotFound)
		return
	}

	if expr.AST != nil && expr.AST.IsLeaf {
		expr.Status = "completed"
		expr.Result = math.Round(expr.AST.Value*100) / 100
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ExprResp{Expression: expr})
}
