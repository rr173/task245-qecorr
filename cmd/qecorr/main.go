// Command qecorr runs the quantum-error-correction syndrome correlation API.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"task245-qecorr/internal/httpapi"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "qecorr.db", "SQLite database path")
	smoke := flag.Bool("smoke-test", false, "run the end-to-end self-check and exit")
	flag.Parse()
	if *smoke {
		if err := service.RunSelfCheck(*dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "smoke-test FAILED: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("smoke-test OK")
		return
	}
	db, err := store.OpenStore(*dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	srv := httpapi.NewServer(service.New(db))
	fmt.Printf("qecorr listening on %s (db=%s)\n", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
