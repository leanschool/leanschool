package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

type RoomHandler struct{ store storage.Storage }

func NewRoomHandler(store storage.Storage) *RoomHandler { return &RoomHandler{store: store} }

func (h *RoomHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /rooms", h.Create)
	mux.HandleFunc("GET /rooms", h.List)
	mux.HandleFunc("GET /rooms/{id}", h.Get)
	mux.HandleFunc("PUT /rooms/{id}", h.Update)
	mux.HandleFunc("DELETE /rooms/{id}", h.Delete)
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "room_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var room model.Room
	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateRoom(r.Context(), &room); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "room_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	rooms, err := h.store.ListRooms(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rooms == nil {
		rooms = []*model.Room{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func (h *RoomHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "room_read") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	room, err := h.store.GetRoom(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if room == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "room_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var room model.Room
	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	room.ID = r.PathValue("id")
	if err := h.store.UpdateRoom(r.Context(), &room); err != nil {
		if errors.Is(err, storage.ErrOptimisticLock) {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasRole(ClaimsFromContext(r.Context()), "room_write") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.store.DeleteRoom(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
