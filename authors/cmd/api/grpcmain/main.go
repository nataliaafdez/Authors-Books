package main

import (
	"log"
	"os"

	"authorsbooks/authors/internal/controller"
	"authorsbooks/authors/internal/grpcserver"
	"authorsbooks/authors/internal/pkg/clients"
	"authorsbooks/authors/internal/repository/memory"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	repo := memory.NewAuthorRepo()

	metadataAddr := getenv("METADATA_ADDR", "metadata-grpc:9092")
	booksAddr := getenv("BOOKS_ADDR", "books-grpc:9091")

	metadataClient := clients.NewMetadataClient(metadataAddr)
	booksClient := clients.NewBooksClient(booksAddr)

	ctrl := controller.NewAuthorController(repo, metadataClient, booksClient)

	addr := getenv("ADDR_GRPC", ":9090")
	log.Printf("[authors grpc] listening on %s", addr)

	if err := grpcserver.Run(ctrl, addr); err != nil {
		log.Fatal(err)
	}
}
