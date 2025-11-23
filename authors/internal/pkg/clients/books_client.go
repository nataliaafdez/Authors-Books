package clients

import (
	"context"
	"fmt"
	"time"

	booksv1 "authorsbooks/proto/booksv1"

	"google.golang.org/grpc"
)

type CreateBookReq struct {
	AuthorID int64
	Title    string
	Year     int
	Genre    string
	Language string
}

type BookResp struct {
	ID       int64
	Title    string
	Year     int
	Genre    string
	Language string
}

type BooksClient struct {
	client booksv1.BooksServiceClient
}

func NewBooksClient(target string) *BooksClient {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithInsecure(), // demo/lab
		grpc.WithBlock(),
	)
	if err != nil {
		panic(fmt.Sprintf("books grpc dial (%s): %v", target, err))
	}

	return &BooksClient{
		client: booksv1.NewBooksServiceClient(conn),
	}
}

func (c *BooksClient) CreateBook(in CreateBookReq) ([]BookResp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --> Aquí llamamos al método gRPC AddBook
	resp, err := c.client.AddBook(ctx, &booksv1.AddBookRequest{
		AuthorId: in.AuthorID,
		Title:    in.Title,
		Year:     int32(in.Year),
		Genre:    in.Genre,
		Language: in.Language,
	})
	if err != nil {
		return nil, fmt.Errorf("books.AddBook grpc error: %w", err)
	}

	out := make([]BookResp, 0, len(resp.Books))
	for _, b := range resp.Books {
		out = append(out, BookResp{
			ID:       b.Id,
			Title:    b.Title,
			Year:     int(b.Year),
			Genre:    b.Genre,
			Language: b.Language,
		})
	}

	return out, nil
}
