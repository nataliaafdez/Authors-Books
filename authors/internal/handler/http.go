package handler

import (
	"authorsbooks/authors/internal/controller"
	"authorsbooks/authors/internal/handler/dto"
	"encoding/json"
	stdhttp "net/http" //hace que no choque con los nombre locales
	"strconv"
	"strings"
)

type Handler struct{ ctrl *controller.AuthorController } //guarda una referencia al controlador

func New(c *controller.AuthorController) *Handler { return &Handler{ctrl: c} } //CONSTRUCTOR:) inyecta el controller

func (h *Handler) Routes(mux *stdhttp.ServeMux) { //registra las rutas en lo de mux
	mux.HandleFunc("/api/v1/authors", h.createAuthor) //crea el autor, acuerdate que el v1 lo puse por las versiones, por si algun dia lo quiero cambiar
	mux.HandleFunc("/api/v1/authors/", h.routeAuthorsSubpaths)
}

func (h *Handler) routeAuthorsSubpaths(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/") //dvide la ruta por "/" y quita "/" al inicio/fin.

	//GET
	if len(parts) == 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "authors" && r.Method == stdhttp.MethodGet {
		h.getAuthor(w, r, parts[3]) //Llama al handler de obtener autor con el id como string.
		return
	}

	//POST
	if len(parts) == 5 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "authors" && parts[4] == "books" && r.Method == stdhttp.MethodPost {
		h.addBookToAuthor(w, r) //llama al handler para agregar un libro a un autor.
		return
	}

	stdhttp.NotFound(w, r) // Si no coincide ningún patrón, responde 404.
}

func (h *Handler) createAuthor(w stdhttp.ResponseWriter, r *stdhttp.Request) { //Crea un autor
	if r.Method != stdhttp.MethodPost { //acepta solo post
		stdhttp.Error(w, "it has to be a post method", stdhttp.StatusMethodNotAllowed)
		return
	}
	var in dto.CreateAuthorReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Name) == "" { //Decodifica y valida nombre.
		stdhttp.Error(w, "invalid body", stdhttp.StatusBadRequest)
		return
	}
	a, err := h.ctrl.CreateAuthor(in.Name) //Llama al controller para crear el autor.
	if err != nil {
		stdhttp.Error(w, err.Error(), stdhttp.StatusBadRequest)
		return
	}
	out := dto.AuthorResp{ID: a.ID, Name: a.Name, Books: make([]dto.BookResp, 0, len(a.Books))}
	for _, b := range a.Books { //normalmente esto está vacío cuando se crea el autor
		out.Books = append(out.Books, dto.BookResp{
			ID: b.ID, Title: b.Title, Year: b.Year, Genre: b.Genre, Language: b.Language, //Mapea dominio -> DTO
		})
	}
	writeJSON(w, stdhttp.StatusCreated, out)
}

func (h *Handler) addBookToAuthor(w stdhttp.ResponseWriter, r *stdhttp.Request) { //agrega un libro a un autor
	if r.Method != stdhttp.MethodPost { //otra vez solo acepta post
		stdhttp.Error(w, "method not allowed", stdhttp.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/") //divide la ruta
	//esto es para que por si el usuario se equivoca, separa lo fijo de lo variable que en este caso lo variable es el ID, verifica que cumple con el formato
	if len(parts) != 5 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "authors" || parts[4] != "books" {
		stdhttp.NotFound(w, r)
		return
	}
	authorID, err := strconv.ParseInt(parts[3], 10, 64) //convierte el id de string a un int
	if err != nil {
		stdhttp.Error(w, "invalid author id", stdhttp.StatusBadRequest)
		return
	}
	var in dto.AddBookReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		stdhttp.Error(w, "invalid body", stdhttp.StatusBadRequest)
		return
	}
	a, err := h.ctrl.AddBook(authorID, in.Title, in.Year, in.Genre, in.Language) //llama al controller para agregar libro
	if err != nil {
		code := stdhttp.StatusBadRequest
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "not found") {
			code = stdhttp.StatusNotFound
		}
		stdhttp.Error(w, err.Error(), code)
		return
	}
	out := dto.AuthorResp{ID: a.ID, Name: a.Name, Books: make([]dto.BookResp, 0, len(a.Books))} //prepara la respuesta
	for _, b := range a.Books {                                                                 //copia los libros actualizados
		out.Books = append(out.Books, dto.BookResp{
			ID: b.ID, Title: b.Title, Year: b.Year, Genre: b.Genre, Language: b.Language,
		})
	}
	writeJSON(w, stdhttp.StatusCreated, out)
}

func (h *Handler) getAuthor(w stdhttp.ResponseWriter, r *stdhttp.Request, idStr string) { //obtiene el autor por id
	authorID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		stdhttp.Error(w, "invalid author id", stdhttp.StatusBadRequest)
		return
	}
	a, err := h.ctrl.GetAuthor(authorID) //llama al controller para leer el autor + libro
	if err != nil {
		code := stdhttp.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			code = stdhttp.StatusNotFound
		}
		stdhttp.Error(w, err.Error(), code)
		return
	}
	out := dto.AuthorResp{ID: a.ID, Name: a.Name, Books: make([]dto.BookResp, 0, len(a.Books))}
	for _, b := range a.Books { //copia los libros del dominio al DTO
		out.Books = append(out.Books, dto.BookResp{
			ID: b.ID, Title: b.Title, Year: b.Year, Genre: b.Genre, Language: b.Language,
		})
	}
	writeJSON(w, stdhttp.StatusOK, out)
}

func writeJSON(w stdhttp.ResponseWriter, code int, v any) { //pregunta si esto esta bien
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
