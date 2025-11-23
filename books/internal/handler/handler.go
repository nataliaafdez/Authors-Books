package handler

import (
	"authorsbooks/books/internal/controller"
	"authorsbooks/books/internal/handler/dto" //es para la entrada y salida del JSON
	"encoding/json"
	stdhttp "net/http"
)

type Handler struct{ ctrl *controller.BookController } //guarda el controller de libros

func New(c *controller.BookController) *Handler { return &Handler{ctrl: c} } //crea el handler y le pasa el controller de libros

// POST /internal/v1/books -> []BookResp
func (h *Handler) CreateForAuthor(w stdhttp.ResponseWriter, r *stdhttp.Request) { //define un endpoint HTTP cuando alguien hace POST se llama a esa función, el chiste de esto es que crea uno o más libros para un autor
	if r.Method != stdhttp.MethodPost { //si el metodo no es post entonces bye
		w.WriteHeader(stdhttp.StatusMethodNotAllowed)
		return
	}
	var in dto.CreateBookReq //decodifica el body JSON a CreateBook
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		return
	}

	books, err := h.ctrl.CreateForAuthor(in.AuthorID, in.Title, in.Year, in.Genre, in.Language) //llama al controller con los datos del libro
	//el controller se encarga de la LOGICA y de comunicarse con el repo
	if err != nil {
		w.WriteHeader(stdhttp.StatusInternalServerError)
		return //deuveule la lista de lubros creada
	}

	out := make([]dto.BookResp, 0, len(books))
	for _, b := range books {
		out = append(out, dto.BookResp{
			ID:       b.ID,
			Title:    b.Title,
			Year:     b.Year,
			Genre:    b.Genre,
			Language: b.Language})
	}
	w.Header().Set("Content-Type", "application/json") //indica que la respuesta será JSON
	_ = json.NewEncoder(w).Encode(out)
}
