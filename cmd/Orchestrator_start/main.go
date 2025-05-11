package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/MrM2025/rpforcalc/tree/master/calc_go/internal/application"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
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
	app.RunOrchestrator()
}
