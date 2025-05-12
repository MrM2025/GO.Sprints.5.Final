package application

import (
	"encoding/json"
	"math"
	"net/http"
	"sync"
)

type SlExprsResp struct {
	Expressions []*Expression `json:"expression,omitempty"`
}

type JWTforExpr struct {
	JWT string `json:"jwt,omitempty"`
}

var EmptyExpression = &Expression{
	Status: "",
}

// TODO: если нет ни одного expression, то выводить сообщение об их отсутствии
func (o *Orchestrator) ExpressionsOutput(w http.ResponseWriter, r *http.Request) { //Сервер, который выводит все переданные серверу выражения
	var (
		mu sync.Mutex
		wt JWTforExpr
	)

	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&wt); err != nil {
		http.Error(w, "Decoding jwt error", http.StatusConflict)
		return
	}

	err := strimJWT("User", wt.JWT)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Session time is up, please, sign in again")
		return
	}

	exprs := make([]*Expression, 0, len(o.ExprStore))

	for _, expr := range o.ExprStore {
		if expr.Jwt != wt.JWT {
			continue
		}

		if expr.AST != nil && expr.AST.IsLeaf {
			expr.Status = "completed"
			expr.Result = math.Round(expr.AST.Value*100) / 100
		}
		exprs = append(exprs, expr)
	}

	if len(exprs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Nothing to post")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SlExprsResp{Expressions: exprs})
}
