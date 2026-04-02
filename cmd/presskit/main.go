package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-presskit/internal/server"
	"github.com/stockyard-dev/stockyard-presskit/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9210"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./presskit-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("presskit: open database: %v", err)
	}
	defer db.Close()

	srv := server.New(db)

	fmt.Printf("\n  Presskit — Self-hosted press kit and media page\n")
	fmt.Printf("  ─────────────────────────────────\n")
	fmt.Printf("  Dashboard:  http://localhost:%s/ui\n", port)
	fmt.Printf("  API:        http://localhost:%s/api\n", port)
	fmt.Printf("  Data:       %s\n", dataDir)
	fmt.Printf("  ─────────────────────────────────\n\n")

	log.Printf("presskit: listening on :%s", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("presskit: %v", err)
	}
}
