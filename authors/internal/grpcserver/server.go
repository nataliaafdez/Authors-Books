package grpcserver

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"authorsbooks/authors/internal/controller"
	authorsv1 "authorsbooks/proto/authorsv1"
)

type Server struct {
	authorsv1.UnimplementedAuthorsServiceServer
	ctrl *controller.AuthorController
}

// constructor
func New(ctrl *controller.AuthorController) *Server {
	return &Server{ctrl: ctrl}
}

func (s *Server) CreateAuthor(ctx context.Context, in *authorsv1.CreateAuthorRequest) (*authorsv1.CreateAuthorResponse, error) {
	awb, err := s.ctrl.CreateAuthor(in.GetName())
	if err != nil {
		return nil, err
	}

	return &authorsv1.CreateAuthorResponse{
		Author: &authorsv1.Author{
			Id:   awb.ID,
			Name: awb.Name,
		},
	}, nil
}

func (s *Server) GetAuthor(ctx context.Context, in *authorsv1.GetAuthorRequest) (*authorsv1.GetAuthorResponse, error) {
	awb, err := s.ctrl.GetAuthor(in.GetId())
	if err != nil {
		return nil, err
	}

	out := &authorsv1.GetAuthorResponse{
		Author: &authorsv1.Author{
			Id:   awb.ID,
			Name: awb.Name,
		},
	}

	for _, b := range awb.Books {
		out.Books = append(out.Books, &authorsv1.Book{
			Id:       b.ID,
			Title:    b.Title,
			Year:     int32(b.Year),
			Genre:    b.Genre,
			Language: b.Language,
		})
	}

	return out, nil
}

func (s *Server) AddBookToAuthor(ctx context.Context, in *authorsv1.AddBookToAuthorRequest) (*authorsv1.AddBookToAuthorResponse, error) {
	awb, err := s.ctrl.AddBook(
		in.GetAuthorId(),
		in.GetTitle(),
		int(in.GetYear()),
		in.GetGenre(),
		in.GetLanguage(),
	)
	if err != nil {
		return nil, err
	}

	out := &authorsv1.AddBookToAuthorResponse{
		AuthorId: awb.ID,
		Name:     awb.Name,
	}

	for _, b := range awb.Books {
		out.Books = append(out.Books, &authorsv1.Book{
			Id:       b.ID,
			Title:    b.Title,
			Year:     int32(b.Year),
			Genre:    b.Genre,
			Language: b.Language,
		})
	}

	return out, nil
}

func Run(ctrl *controller.AuthorController, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	gs := grpc.NewServer()

	srv := New(ctrl)
	authorsv1.RegisterAuthorsServiceServer(gs, srv)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)

	reflection.Register(gs)

	return gs.Serve(lis)
}
