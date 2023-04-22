package main

import (
	"database/sql"
	"log"
	"mailing/grpcapi"
	"mailing/jsonapi"
	"mailing/mdb"
	"sync"

	"github.com/alexflint/go-arg"
)

var args struct {
	DBPath   string `arg:"env:MAILINGLIST_DB`
	BindJson string `arg:"env:MAILINGLIST_BIND_JSON`
	BindGrpc string `arg:"env:MAILINGLIST_BIND_GRPC`
}

func main() {
	arg.MustParse(&args)
	if args.DBPath == "" {
		args.DBPath = "list.db"
	}
	if args.BindJson == "" {
		args.BindJson = ":8080"
	}
	if args.BindGrpc == "" {
		args.BindGrpc = ":8081"
	}
	log.Printf("using database '%v'\n", args.DBPath)
	db, err := sql.Open("sqlite3", args.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mdb.TryCreate(db)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Printf("starting JSON API server...\n")
		jsonapi.Serve(db, args.BindJson)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		log.Printf("starting gRPC API server...\n")
		grpcapi.Serve(db, args.BindGrpc)
		wg.Done()
	}()
	wg.Wait()
}
