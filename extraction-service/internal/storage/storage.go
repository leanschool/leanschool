package storage

import (
	"github.com/Joel-Haeberli/extraction-service/internal/model"
)

type Storage interface {
	CreateTemplate(template model.Template) (model.Template, error)
	UpdateTemplate(template model.Template) (model.Template, error)
	DeleteTemplate(id string) error
	GetTemplate(id string) (model.Template, error)
	GetAllTemplates() ([]model.Template, error)
}
