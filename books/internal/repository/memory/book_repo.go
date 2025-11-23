package memory

import (
	"authorsbooks/books/internal/domain"
	"sync"
	"sync/atomic" //incrementar id de manera segura
)

type Repo struct { //es como una base de datos en ram
	mu       sync.RWMutex            //candado de lectura solamente
	nextID   int64                   //asigna los id
	byAuthor map[int64][]domain.Book //el mapa que es {id, lista de libros}
}

func NewBookRepo() *Repo { //constructor, crea un nuevo repo vacío, inicializa el mapa de libros por autor
	return &Repo{byAuthor: map[int64][]domain.Book{}}
}

func (r *Repo) Create(authorID int64, b domain.Book) ([]domain.Book, error) { //crea un libro para un autor y devuelve la lista actualizada de los libros de ese autor
	r.mu.Lock()
	defer r.mu.Unlock()
	b.ID = atomic.AddInt64(&r.nextID, 1) //Asigna un ID único al libro, incrementando el nextid

	r.byAuthor[authorID] = append(r.byAuthor[authorID], b) //Agrega el libro al slice de ese autor

	out := make([]domain.Book, len(r.byAuthor[authorID])) //rea un slice nuevo para devolver (copia segura).

	copy(out, r.byAuthor[authorID]) //Copia los libros actuales del autor a "out".
	return out, nil                 //devuevlve la lista de libros del autor
}
