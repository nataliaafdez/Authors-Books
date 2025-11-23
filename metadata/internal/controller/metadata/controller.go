package metadata

import (
	"context"

	"authorsbooks/metadata/internal/repository"
	"authorsbooks/metadata/pkg/model"
)

type metadataRepository interface { //define como debe comportarse cada repo del metadata
	Get(ctx context.Context, id string) (*model.Metadata, error) //obetenr metadata con un ID
	Put(ctx context.Context, id string, m *model.Metadata) error //guardar metadata con  un ID
}

type Controller struct {
	repo metadataRepository
}

func New(repo metadataRepository) *Controller {
	return &Controller{repo: repo}
}

// para leer metadata
func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	m, err := c.repo.Get(ctx, id) //lama al repo para obtener metadata por id
	if err != nil {               //se hubo error entonces hay un "no econtrado"
		return nil, repository.ErrNotFound
	}
	return m, nil
}

// para guardar metadata nueva o actualizar la qu ya exista
func (c *Controller) Put(ctx context.Context, m *model.Metadata) error {
	if m == nil || m.ID == "" { //no se puede si no hay un objeto
		return repository.ErrNotFound
	}
	return c.repo.Put(ctx, m.ID, m) //llama al repo para actualizar la metadata
}
