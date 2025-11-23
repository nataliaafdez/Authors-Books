package memory

import (
	"authorsbooks/authors/internal/domain"
	"authorsbooks/authors/internal/repository"
	"errors"
	"sync"
)

type repo struct {
	mu      sync.RWMutex
	nextID  int64
	authors map[int64]domain.Author
	books   map[int64][]domain.Book // libros por authorID
}

func NewAuthorRepo() repository.AuthorRepository { //crea un nuevo repo en memoria
	return &repo{
		nextID:  1,
		authors: make(map[int64]domain.Author),
		books:   make(map[int64][]domain.Book),
	}
}

func (r *repo) Create(name string) (*domain.Author, error) { //crea un nuevo autor en el repo
	r.mu.Lock() //bloquea el modo escritura
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	a := domain.Author{ID: id, Name: name} //crea el autor
	r.authors[id] = a                      //lo guarda en el mapa de autores
	return &a, nil                         //devuelve el autor recien creado
}

func (r *repo) GetByID(id int64) (*domain.Author, error) { //busca un autor por su ID
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.authors[id]
	if !ok {
		return nil, errors.New("author not found")
	}
	return &a, nil
}

func (r *repo) AddBook(authorID int64, b domain.Book) error { //agrega un libro al autor
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.authors[authorID]; !ok {
		return errors.New("author not found")
	}
	r.books[authorID] = append(r.books[authorID], b)
	return nil
}

func (r *repo) ListBooks(authorID int64) ([]domain.Book, error) { //devuelve todos los libros de un autor
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.authors[authorID]; !ok {
		return nil, errors.New("author not found")
	}
	out := make([]domain.Book, len(r.books[authorID]))
	copy(out, r.books[authorID])
	return out, nil
}
