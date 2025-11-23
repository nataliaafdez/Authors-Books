package main

import (
	"log"
	"os"

	"authorsbooks/metadata/internal/controller/metadata"
	"authorsbooks/metadata/internal/grpcserver"
	"authorsbooks/metadata/internal/repository/memory"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	repo := memory.New()
	ctrl := metadata.New(repo)

	addr := getenv("ADDR_GRPC", ":9092")
	log.Printf("[metadata grpc] listening on %s", addr)

	if err := grpcserver.Run(ctrl, addr); err != nil {
		log.Fatal(err)
	}
}
