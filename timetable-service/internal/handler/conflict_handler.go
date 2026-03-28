package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// ConflictHandler provides read-only access to detected conflicts.
type ConflictHandler struct{ store storage.Storage }

func NewConflictHandler(store storage.Storage) *ConflictHandler {
	return &ConflictHandler{store: store}
}

func (h *ConflictHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /plans/{planId}/conflicts", h.List)
}

func (h *ConflictHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	conflicts, err := h.store.ListConflictsByPlan(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if conflicts == nil {
		conflicts = []*model.Conflict{}
	}

	// Optional query param filters
	resolvedFilter := r.URL.Query().Get("resolved")
	teacherFilter := r.URL.Query().Get("teacherId")
	if resolvedFilter != "" || teacherFilter != "" {
		var filtered []*model.Conflict
		for _, c := range conflicts {
			if resolvedFilter == "true" && !c.Resolved {
				continue
			}
			if resolvedFilter == "false" && c.Resolved {
				continue
			}
			if teacherFilter != "" && c.TeacherID != teacherFilter {
				continue
			}
			filtered = append(filtered, c)
		}
		if filtered == nil {
			filtered = []*model.Conflict{}
		}
		conflicts = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conflicts)
}
