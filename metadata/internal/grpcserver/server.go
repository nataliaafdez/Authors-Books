package grpcserver

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"authorsbooks/metadata/internal/controller/metadata"
	"authorsbooks/metadata/pkg/model"
	metadatav1 "authorsbooks/proto/metadatav1"
)

type Server struct {
	metadatav1.UnimplementedMetadataServiceServer
	ctrl *metadata.Controller
}

func New(ctrl *metadata.Controller) *Server { return &Server{ctrl: ctrl} }

func (s *Server) GetMetadata(ctx context.Context, in *metadatav1.GetMetadataRequest) (*metadatav1.GetMetadataResponse, error) {
	m, err := s.ctrl.Get(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	return &metadatav1.GetMetadataResponse{
		Id:          m.ID,
		Title:       m.Title,
		Description: m.Description,
	}, nil
}

func (s *Server) PutMetadata(ctx context.Context, in *metadatav1.PutMetadataRequest) (*metadatav1.PutMetadataResponse, error) {
	md := &model.Metadata{
		ID:          in.GetId(),
		Title:       in.GetTitle(),
		Description: in.GetDescription(),
	}
	if err := s.ctrl.Put(ctx, md); err != nil {
		return nil, err
	}
	return &metadatav1.PutMetadataResponse{}, nil
}

func Run(ctrl *metadata.Controller, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	gs := grpc.NewServer()
	metadatav1.RegisterMetadataServiceServer(gs, New(ctrl))

	// health-check
	hs := health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)

	reflection.Register(gs)

	return gs.Serve(lis)
}
