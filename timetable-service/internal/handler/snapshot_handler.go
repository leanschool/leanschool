package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// SnapshotHandler provides read-only access to plan-local snapshots.
type SnapshotHandler struct{ store storage.Storage }

func NewSnapshotHandler(store storage.Storage) *SnapshotHandler {
	return &SnapshotHandler{store: store}
}

func (h *SnapshotHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /plans/{planId}/teachers", h.ListTeachers)
	mux.HandleFunc("GET /plans/{planId}/subjects", h.ListSubjects)
	mux.HandleFunc("GET /plans/{planId}/classes", h.ListClasses)
	mux.HandleFunc("GET /plans/{planId}/rooms", h.ListRooms)
}

func (h *SnapshotHandler) ListTeachers(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	snapshots, err := h.store.ListTeacherSnapshots(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if snapshots == nil {
		snapshots = []*model.TeacherSnapshot{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}

func (h *SnapshotHandler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	snapshots, err := h.store.ListSubjectSnapshots(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if snapshots == nil {
		snapshots = []*model.SubjectSnapshot{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}

func (h *SnapshotHandler) ListClasses(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	snapshots, err := h.store.ListSchoolClassSnapshots(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if snapshots == nil {
		snapshots = []*model.SchoolClassSnapshot{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}

func (h *SnapshotHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	snapshots, err := h.store.ListRoomSnapshots(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if snapshots == nil {
		snapshots = []*model.RoomSnapshot{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}
