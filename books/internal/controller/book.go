package controller

import (
	"authorsbooks/books/internal/domain"
)

type BookRepo interface { //interfaz que define lo que debe hacer el repo de lirbos
	Create(authorID int64, b domain.Book) ([]domain.Book, error) //crea un libro para un autor
}

type BookController struct {
	repo BookRepo //el controller guarda una REFERENCIA a un bookrepo = donde estoy almacenando los libros
}

func NewBookController(r BookRepo) *BookController { //constructor del controller de libros
	return &BookController{repo: r} //devuelve un nuevo book controller con el repo inyectado
}

func (c *BookController) CreateForAuthor(authorID int64, title string, year int, genre, language string) ([]domain.Book, error) { //método del controlador para crar un libro de un autor
	return c.repo.Create(authorID, domain.Book{ //llama al repositorio, armando un objeto libro
		Title:    title,
		Year:     year,
		Genre:    genre,
		Language: language})
}
