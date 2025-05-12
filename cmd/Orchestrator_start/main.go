package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/MrM2025/rpforcalc/tree/master/calc_go/internal/application"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var (
		idnum int
	)
	ctx := context.TODO()

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	app := application.NewOrchestrator(db, ctx)
	app.CreateTables()
	rows, err := db.QueryContext(
		app.Ctx,
		`SELECT id, expression, jwt, user_lg, status, result FROM expressions`,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var expr *application.Expression = &application.Expression{}
		if err = rows.Scan(
			&expr.ID,
			&expr.Expr,
			&expr.Jwt,
			&expr.Login,
			&expr.Status,
			&expr.Result,
		); err != nil {
			log.Fatal(err)
		}

		app.ExprStore[expr.ID] = expr

		if expr.Status != "completed" {
			app.Tasks(expr)
		}

	}

	if err = app.Db.QueryRowContext(
		app.Ctx,
		`SELECT COUNT(id) FROM expressions`,
	).Scan(&idnum); err != nil {
		log.Fatal(err)
	}

	app.ExprCounter = idnum

	app.RunOrchestrator()
}
