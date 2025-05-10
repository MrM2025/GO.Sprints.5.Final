package application

import (
	"encoding/json"
	"math"
	"net/http"
	"sync"
)

type JWTforExpr struct {
	JWT string `json:"jwt,omitempty"`
}

var EmptyExpression = &Expression{
	Status: "",
}

// TODO: если нет ни одного expression, то выводить сообщение об их отсутствии
func ExpressionsOutput(w http.ResponseWriter, r *http.Request) { //Сервер, который выводит все переданные серверу выражения
	var (
		mu sync.Mutex
		ut JWTforExpr
	)

	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&ut); err != nil {
		http.Error(w, "Decoding jtw error", http.StatusConflict)
		return
	}

	exprs := make([]*Expression, 0, len(exprStore))

	for _, expr := range exprStore {
		if expr.Jwt != ut.JWT {
			continue
		}

		if expr.AST != nil && expr.AST.IsLeaf {
			expr.Status = "completed"
			expr.Result = math.Round(expr.AST.Value*100) / 100
		}
		exprs = append(exprs, expr)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": exprs})
}
