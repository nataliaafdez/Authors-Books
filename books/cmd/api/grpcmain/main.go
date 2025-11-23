package main

import (
	"log"
	"os"

	"authorsbooks/books/internal/controller"
	"authorsbooks/books/internal/grpcserver"
	"authorsbooks/books/internal/repository/memory"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	repo := memory.NewBookRepo()
	ctrl := controller.NewBookController(repo)

	addr := getenv("ADDR_GRPC", ":9091")
	log.Printf("[books grpc] listening on %s", addr)

	if err := grpcserver.Run(ctrl, addr); err != nil {
		log.Fatal(err)
	}
}
