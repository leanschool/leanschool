package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Joel-Haeberli/extraction-service/internal/model"
	"github.com/Joel-Haeberli/extraction-service/internal/storage"
)

type Handler struct {
	storage    storage.Storage
	processor  *ExtractionProcessor
}

func New(storage storage.Storage) *Handler {
	return &Handler{
		storage:   storage,
		processor: NewExtractionProcessor(),
	}
}

func (h *Handler) Extract(w http.ResponseWriter, r *http.Request) {
	var extraction model.ExtractionTemplate
	if err := json.NewDecoder(r.Body).Decode(&extraction); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.processor.Process(extraction)
	if err != nil {
		http.Error(w, "extraction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var template model.Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.storage.CreateTemplate(template)
	if err != nil {
		http.Error(w, "failed to create template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	var template model.Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.storage.UpdateTemplate(template)
	if err != nil {
		if err.Error() == "template not found" {
			http.Error(w, "template not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to update template: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.DeleteTemplate(id); err != nil {
		if err.Error() == "template not found" {
			http.Error(w, "template not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to delete template: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetAllTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.storage.GetAllTemplates()
	if err != nil {
		http.Error(w, "failed to get templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(templates)
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	template, err := h.storage.GetTemplate(id)
	if err != nil {
		if err.Error() == "template not found" {
			http.Error(w, "template not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to get template: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(template)
}
