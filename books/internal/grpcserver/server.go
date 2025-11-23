package grpcserver

import (
	"context"
	"fmt"
	"net"

	booksv1 "authorsbooks/proto/booksv1"
	"authorsbooks/books/internal/controller"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	booksv1.UnimplementedBooksServiceServer
	ctrl *controller.BookController
}

func New(ctrl *controller.BookController) *Server { return &Server{ctrl: ctrl} }

// Asume: ctrl.CreateForAuthor(authorID, title, year, genre, language) ([]domain.Book, error)
func (s *Server) AddBook(ctx context.Context, in *booksv1.AddBookRequest) (*booksv1.AddBookResponse, error) {
	list, err := s.ctrl.CreateForAuthor(
		in.GetAuthorId(),
		in.GetTitle(),
		int(in.GetYear()),
		in.GetGenre(),
		in.GetLanguage(),
	)
	if err != nil {
		return nil, err
	}

	out := make([]*booksv1.Book, 0, len(list))
	for _, b := range list {
		// AuthorId: si tu dominio no tiene AuthorId, enviamos el de la request.
		authorID := in.GetAuthorId()
		// Si tu dominio sí tiene AuthorId (b.AuthorId != 0), úsalo:
		type hasAuthorID interface{ GetAuthorId() int64 } // nota: solo heurístico si usas getters
		_ = hasAuthorID(nil)                              // silencia import no usado si quitas esto

		out = append(out, &booksv1.Book{
			Id:       b.ID,
			AuthorId: authorID,
			Title:    b.Title,
			Year:     int32(b.Year),
			Genre:    b.Genre,
			Language: b.Language,
		})
	}

	return &booksv1.AddBookResponse{Books: out}, nil
}

func Run(ctrl *controller.BookController, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil { return fmt.Errorf("listen: %w", err) }

	gs := grpc.NewServer()
	booksv1.RegisterBooksServiceServer(gs, New(ctrl))

	// Health + Reflection para grpcurl list
	hs := health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)
	reflection.Register(gs)

	return gs.Serve(lis)
}
