package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// EntryHandler handles CRUD and operations for TimetableEntry.
type EntryHandler struct{ store storage.Storage }

func NewEntryHandler(store storage.Storage) *EntryHandler {
	return &EntryHandler{store: store}
}

func (h *EntryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /plans/{planId}/entries", h.List)
	mux.HandleFunc("GET /plans/{planId}/entries/{id}", h.Get)
	mux.HandleFunc("PUT /plans/{planId}/entries/{id}", h.Update)
	mux.HandleFunc("POST /plans/{planId}/entries/{id}/swap", h.Swap)
	mux.HandleFunc("POST /plans/{planId}/entries/{id}/reassign", h.Reassign)
}

func (h *EntryHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	entries, err := h.store.ListEntriesByPlan(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []*model.TimetableEntry{}
	}

	// Optional query param filters
	classID := r.URL.Query().Get("classId")
	teacherID := r.URL.Query().Get("teacherId")
	if classID != "" || teacherID != "" {
		var filtered []*model.TimetableEntry
		for _, e := range entries {
			if classID != "" && e.SchoolClassID != classID {
				continue
			}
			if teacherID != "" && e.TeacherID != teacherID {
				continue
			}
			filtered = append(filtered, e)
		}
		if filtered == nil {
			filtered = []*model.TimetableEntry{}
		}
		entries = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *EntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	entry, err := h.store.GetEntry(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entry == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (h *EntryHandler) requireResolvingStatus(w http.ResponseWriter, r *http.Request) bool {
	plan, err := h.store.GetPlan(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	if plan == nil {
		http.Error(w, "plan not found", http.StatusNotFound)
		return false
	}
	if plan.Status != model.PlanStatusResolving {
		http.Error(w, "entries can only be modified when plan is in resolving status", http.StatusBadRequest)
		return false
	}
	return true
}

func (h *EntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "timetable_resolve") && !hasRole(claims, "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !h.requireResolvingStatus(w, r) {
		return
	}
	var entry model.TimetableEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	entry.ID = r.PathValue("id")
	entry.PlanID = r.PathValue("planId")
	if err := h.store.UpdateEntry(r.Context(), &entry); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// Swap exchanges the time slots of two timetable entries.
func (h *EntryHandler) Swap(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "timetable_resolve") && !hasRole(claims, "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !h.requireResolvingStatus(w, r) {
		return
	}
	var req model.SwapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	sourceID := r.PathValue("id")
	source, err := h.store.GetEntry(r.Context(), sourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if source == nil {
		http.Error(w, "source entry not found", http.StatusNotFound)
		return
	}

	target, err := h.store.GetEntry(r.Context(), req.TargetEntryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if target == nil {
		http.Error(w, "target entry not found", http.StatusNotFound)
		return
	}

	// Swap time slot IDs
	source.TimeSlotID, target.TimeSlotID = target.TimeSlotID, source.TimeSlotID

	if err := h.store.UpdateEntry(r.Context(), source); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.UpdateEntry(r.Context(), target); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"source": source,
		"target": target,
	})
}

// Reassign updates the teacher assigned to an entry.
func (h *EntryHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if !hasRole(claims, "timetable_resolve") && !hasRole(claims, "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !h.requireResolvingStatus(w, r) {
		return
	}
	var req model.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	entry, err := h.store.GetEntry(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entry == nil {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	entry.TeacherID = req.TeacherID
	if err := h.store.UpdateEntry(r.Context(), entry); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}
