package main

import (
	"database/sql"
	"log"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/naufalihsan/mailer/api"
	mdb "github.com/naufalihsan/mailer/db"
)

var args struct {
	DbPath   string `arg:"env:MAILER_DB"`
	BindJson string `arg:"env:MAILER_BIND_JSON"`
	BindGrpc string `arg:"env:MAILER_BIND_GRPC"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "mailer.db"
	}

	if args.BindJson == "" {
		args.BindJson = ":8080"
	}

	if args.BindGrpc == "" {
		args.BindGrpc = ":8081"
	}

	log.Printf("using db '%v'", args.DbPath)

	db, err := sql.Open("sqlite3", args.DbPath)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	mdb.CreateDB(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Println("starting REST server ...")
		api.ServeREST(db, args.BindJson)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		log.Println("starting gRPC server ...")
		api.ServeRPC(db, args.BindGrpc)
		wg.Done()
	}()

	wg.Wait()
}
