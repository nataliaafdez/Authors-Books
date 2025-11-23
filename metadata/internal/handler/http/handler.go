package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"authorsbooks/metadata/internal/controller/metadata"
	"authorsbooks/metadata/internal/repository"
	"authorsbooks/metadata/pkg/model"
)

type Handler struct {
	ctrl *metadata.Controller //el handler necesita del controller para acceder a la lógica del negocio
}

func New(ctrl *metadata.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

// Solo acepta GET , si no es get BYE
func (h *Handler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id") //Lee el parámetro del id de la URL, si no vino error
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	m, err := h.ctrl.Get(r.Context(), id) //llama al controller para obtner la metadata por id
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("get error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(m) //si todo OK entonces responde con la metadata en JSON
}

// SOlo acepta POST
func (h *Handler) PostMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in model.Metadata                                                      //prepara un objeto metadata
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.ID == "" { //decodifica el json recibido al structr del metatda
		http.Error(w, "invalid body (need id,title,description,director)", http.StatusBadRequest)
		return
	}

	if err := h.ctrl.Put(r.Context(), &in); err != nil { //llama al controller para guardar la metadata
		log.Printf("put error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
