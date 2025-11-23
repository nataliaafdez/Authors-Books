package repository

import "authorsbooks/authors/internal/domain"

type AuthorRepository interface {
	Create(name string) (*domain.Author, error)
	GetByID(id int64) (*domain.Author, error)

	AddBook(authorID int64, b domain.Book) error
	ListBooks(authorID int64) ([]domain.Book, error)
}
