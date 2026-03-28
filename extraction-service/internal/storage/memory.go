package storage

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"

	"github.com/Joel-Haeberli/extraction-service/internal/model"
)

type MemoryStorage struct {
	mu        sync.RWMutex
	templates map[string]model.Template
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		templates: make(map[string]model.Template),
	}
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (s *MemoryStorage) CreateTemplate(template model.Template) (model.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if template.Id == "" {
		template.Id = newID()
	}

	s.templates[template.Id] = template
	return template, nil
}

func (s *MemoryStorage) UpdateTemplate(template model.Template) (model.Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.templates[template.Id]; !exists {
		return model.Template{}, errors.New("template not found")
	}

	s.templates[template.Id] = template
	return template, nil
}

func (s *MemoryStorage) DeleteTemplate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.templates[id]; !exists {
		return errors.New("template not found")
	}

	delete(s.templates, id)
	return nil
}

func (s *MemoryStorage) GetTemplate(id string) (model.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	template, exists := s.templates[id]
	if !exists {
		return model.Template{}, errors.New("template not found")
	}

	return template, nil
}

func (s *MemoryStorage) GetAllTemplates() ([]model.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]model.Template, 0, len(s.templates))
	for _, template := range s.templates {
		templates = append(templates, template)
	}

	return templates, nil
}
