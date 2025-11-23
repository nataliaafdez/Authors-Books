package controller

import (
	"authorsbooks/authors/internal/domain"
	"authorsbooks/authors/internal/pkg/clients"
	"authorsbooks/authors/internal/repository"
	"errors"
	"strings"
)

type AuthorController struct { //esta es la estructura del controller de authors
	repo repository.AuthorRepository //dependencia al repositorio de guardar / leer autores y libros
	//es una interfaz que en go es que no se copia toda la implementacipn si no que solo copia la referncia
	meta  *clients.MetadataClient //consulta el catalogo en metadara
	books *clients.BooksClient    //crea libros en el microservicio de books
}

// CONSTRUCTOR :)
// crear un autor nuevo, valida el nombre y crea un autor en el repo, sirve para dar de "alta" los autores y devuelve al autor con su lista de libros que esta vacia
func NewAuthorController(r repository.AuthorRepository, meta *clients.MetadataClient, books *clients.BooksClient) *AuthorController {
	return &AuthorController{repo: r, meta: meta, books: books} //inyecta dependencias y devuelve el controller
}

func (c *AuthorController) CreateAuthor(name string) (*domain.AuthorWithBooks, error) { //crea un autor nuevo
	name = strings.TrimSpace(name) //limpia los espacios
	if name == "" {                //si el nombre no está vacío
		return nil, errors.New("name required")
	}
	a, err := c.repo.Create(name) //pide que el repo cree el autor
	if err != nil {
		return nil, err
	}
	return &domain.AuthorWithBooks{ //devuelve el autor con su lista de libros vacía
		ID:    a.ID,            //id que es generado por el repo
		Name:  a.Name,          //nombre del autor
		Books: []domain.Book{}, //lista de libros sin libros hasta que se llenen
	}, nil
}

func (c *AuthorController) AddBook(authorID int64, title string, year int, genre, language string) (*domain.AuthorWithBooks, error) { //agrega un libro a un autor
	title = strings.TrimSpace(title)
	genre = strings.TrimSpace(genre)
	language = strings.TrimSpace(language)
	if title == "" {
		return nil, errors.New("title required")
	}

	if cat, err := c.meta.GetCatalog(); err == nil { //intenta obtener catalogo desde metadata
		if genre != "" && !containsFold(cat.Genres, genre) { //si se indico genero y no esta en el catalogo entonces manda error
			return nil, errors.New("invalid genre")
		}
		if language != "" && !containsFold(cat.Languages, language) { //lo mismo pero con el idioma
			return nil, errors.New("invalid language")
		}
	}

	// Verifica que exista el autor
	a, err := c.repo.GetByID(authorID) //consulta al autor en el repo
	if err != nil {                    //si no existe o falla entonces error
		return nil, err
	}

	blist, err := c.books.CreateBook(clients.CreateBookReq{ //INVOCA jaja a books para crear el libro
		AuthorID: authorID,
		Title:    title,
		Year:     year,
		Genre:    genre,
		Language: language,
	})
	if err != nil {
		return nil, err
	}

	for _, b := range blist { //recorre los libros creados
		_ = c.repo.AddBook(authorID, domain.Book{ //los guarda en el repo local
			ID:       b.ID,
			Title:    b.Title,
			Year:     b.Year,
			Genre:    b.Genre,
			Language: b.Language,
		})
	}

	// Responder lo que quedó persistido
	all, _ := c.repo.ListBooks(authorID)                                    //obtiene todos los libros del autor
	return &domain.AuthorWithBooks{ID: a.ID, Name: a.Name, Books: all}, nil //devuelve al autor con su lista de libros acturalizada
}

func (c *AuthorController) GetAuthor(authorID int64) (*domain.AuthorWithBooks, error) { //obtiene un autor con sus libros
	a, err := c.repo.GetByID(authorID) //busca el autor en el repo, si no falla,

	if err != nil {
		return nil, err
	}

	books, _ := c.repo.ListBooks(authorID)                                    //consultar los libros
	return &domain.AuthorWithBooks{ID: a.ID, Name: a.Name, Books: books}, nil //te devuelve el autor + sus libros
}

func containsFold(list []string, v string) bool { //compribea si v esta en list ignroadndo las cosas que se pueden escribir mal como mayus, minus y eso
	v = strings.ToLower(strings.TrimSpace(v))
	for _, x := range list {
		if strings.ToLower(strings.TrimSpace(x)) == v {
			return true
		}
	}
	return false
}
