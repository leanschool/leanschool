package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Joel-Haeberli/timetable-service/internal/model"
	"github.com/Joel-Haeberli/timetable-service/internal/storage"
)

// TimeSlotHandler handles CRUD for TimeSlotDefinition.
type TimeSlotHandler struct{ store storage.Storage }

func NewTimeSlotHandler(store storage.Storage) *TimeSlotHandler {
	return &TimeSlotHandler{store: store}
}

func (h *TimeSlotHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /plans/{planId}/time-slots", h.Create)
	mux.HandleFunc("GET /plans/{planId}/time-slots", h.List)
	mux.HandleFunc("PUT /plans/{planId}/time-slots/{id}", h.Update)
	mux.HandleFunc("DELETE /plans/{planId}/time-slots/{id}", h.Delete)
	mux.HandleFunc("POST /plans/{planId}/time-slots/generate-default", h.GenerateDefault)
}

func (h *TimeSlotHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var ts model.TimeSlotDefinition
	if err := json.NewDecoder(r.Body).Decode(&ts); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ts.PlanID = r.PathValue("planId")
	if err := h.store.CreateTimeSlot(r.Context(), &ts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ts)
}

func (h *TimeSlotHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	slots, err := h.store.ListTimeSlotsByPlan(r.Context(), r.PathValue("planId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if slots == nil {
		slots = []*model.TimeSlotDefinition{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}

func (h *TimeSlotHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var ts model.TimeSlotDefinition
	if err := json.NewDecoder(r.Body).Decode(&ts); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ts.ID = r.PathValue("id")
	ts.PlanID = r.PathValue("planId")
	if err := h.store.UpdateTimeSlot(r.Context(), &ts); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ts)
}

func (h *TimeSlotHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteTimeSlot(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GenerateDefault creates a standard Mon-Fri time slot grid based on the request params.
func (h *TimeSlotHandler) GenerateDefault(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "timetable_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req model.GenerateDefaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	planID := r.PathValue("planId")

	// Delete existing slots for this plan first
	if err := h.store.DeleteTimeSlotsByPlan(r.Context(), planID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse start time
	start, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		http.Error(w, "invalid startTime format, expected HH:MM: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.LessonDurationMin <= 0 {
		req.LessonDurationMin = 45
	}
	if req.BreakDurationMin <= 0 {
		req.BreakDurationMin = 5
	}
	if req.LunchBreakMin <= 0 {
		req.LunchBreakMin = 60
	}

	var created []*model.TimeSlotDefinition
	for day := model.Monday; day <= model.Friday; day++ {
		cursor := start
		period := 1

		// Morning periods
		for i := 0; i < req.MorningPeriods; i++ {
			end := cursor.Add(time.Duration(req.LessonDurationMin) * time.Minute)
			ts := &model.TimeSlotDefinition{
				PlanID:    planID,
				DayOfWeek: day,
				Period:    period,
				StartTime: cursor.Format("15:04"),
				EndTime:   end.Format("15:04"),
				IsMorning: true,
			}
			if err := h.store.CreateTimeSlot(r.Context(), ts); err != nil {
				http.Error(w, fmt.Sprintf("creating slot: %v", err), http.StatusInternalServerError)
				return
			}
			created = append(created, ts)
			period++
			cursor = end.Add(time.Duration(req.BreakDurationMin) * time.Minute)
		}

		// Lunch break
		cursor = cursor.Add(time.Duration(req.LunchBreakMin-req.BreakDurationMin) * time.Minute)

		// Afternoon periods
		for i := 0; i < req.AfternoonPeriods; i++ {
			end := cursor.Add(time.Duration(req.LessonDurationMin) * time.Minute)
			ts := &model.TimeSlotDefinition{
				PlanID:    planID,
				DayOfWeek: day,
				Period:    period,
				StartTime: cursor.Format("15:04"),
				EndTime:   end.Format("15:04"),
				IsMorning: false,
			}
			if err := h.store.CreateTimeSlot(r.Context(), ts); err != nil {
				http.Error(w, fmt.Sprintf("creating slot: %v", err), http.StatusInternalServerError)
				return
			}
			created = append(created, ts)
			period++
			cursor = end.Add(time.Duration(req.BreakDurationMin) * time.Minute)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// extractPlanID parses the planId from a path like /plans/{planId}/...
// This is unused since Go 1.22 added PathValue, but kept for documentation.
func extractPlanID(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
