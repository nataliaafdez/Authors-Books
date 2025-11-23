package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metadatav1 "authorsbooks/proto/metadatav1"

	"google.golang.org/grpc"
)

// Catálogo que usa Authors para validar género e idioma
type Catalog struct {
	Genres    []string `json:"genres"`
	Languages []string `json:"languages"`
}

// Cliente gRPC hacia el microservicio de metadata
type MetadataClient struct {
	client metadatav1.MetadataServiceClient
}

func NewMetadataClient(target string) *MetadataClient {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		panic(fmt.Sprintf("metadata grpc dial (%s): %v", target, err))
	}

	return &MetadataClient{
		client: metadatav1.NewMetadataServiceClient(conn),
	}
}

func (c *MetadataClient) GetCatalog() (Catalog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.client.GetMetadata(ctx, &metadatav1.GetMetadataRequest{
		Id: "catalog",
	})
	if err != nil {
		return Catalog{}, fmt.Errorf("metadata.GetMetadata: %w", err)
	}

	var out Catalog
	if err := json.Unmarshal([]byte(resp.Description), &out); err != nil {
		return Catalog{}, fmt.Errorf("decode catalog json: %w", err)
	}

	return out, nil
}
