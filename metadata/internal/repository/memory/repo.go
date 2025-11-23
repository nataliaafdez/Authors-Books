package memory

import (
	"context"
	"sync"

	"authorsbooks/metadata/internal/repository"
	"authorsbooks/metadata/pkg/model"
)

type Repository struct { //base de daros en RAM
	sync.RWMutex
	data map[string]*model.Metadata
}

func New() *Repository {
	return &Repository{
		data: map[string]*model.Metadata{
			"1": {ID: "1", Title: "A Court of thorns and roses", Description: "Ejemplo ejemplo metadata"},
		},
	}
} //devuelve un repo vacío con un regisyto de ejemplo

func (r *Repository) Get(_ context.Context, id string) (*model.Metadata, error) { //obtiene metadata por ID
	r.RLock()
	defer r.RUnlock()
	m, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return m, nil
}

func (r *Repository) Put(_ context.Context, id string, metadata *model.Metadata) error { //inserta o actualiza metadata por ID
	r.Lock()
	defer r.Unlock()
	r.data[id] = metadata
	return nil
}
