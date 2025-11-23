package domain

type Author struct { //autor
	ID   int64  `json:"id"`
	Name string `json:"name"` //solo le voy a poner el nombre así, ya despues si se lo quiero cambiar pues ya veré
}

type Book struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
	Language string `json:"language"`
}
type AuthorWithBooks struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Books []Book `json:"books"`
}
