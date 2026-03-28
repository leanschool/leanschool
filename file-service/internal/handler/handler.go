package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadSize = 20 << 20 // 20 MB

// Handler serves file upload, download, and delete operations.
type Handler struct {
	dataDir string
}

func New(dataDir string) *Handler {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("[file-service] creating data dir %s: %v", dataDir, err)
	}
	return &Handler{dataDir: dataDir}
}

// Upload handles POST /files
// Expects multipart/form-data with a "file" field.
// Returns {"id": "...", "contentType": "..."}.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "request too large or malformed", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	id := newID()

	dst, err := os.Create(filepath.Join(h.dataDir, id))
	if err != nil {
		http.Error(w, "storing file failed", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	n, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(filepath.Join(h.dataDir, id))
		http.Error(w, "writing file failed", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(filepath.Join(h.dataDir, id+".ct"), []byte(contentType), 0644); err != nil {
		os.Remove(filepath.Join(h.dataDir, id))
		http.Error(w, "storing metadata failed", http.StatusInternalServerError)
		return
	}

	log.Printf("[file-service] stored id=%s size=%d contentType=%s", id, n, contentType)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "contentType": contentType})
}

// Download handles GET /files/{id}
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	ctBytes, err := os.ReadFile(filepath.Join(h.dataDir, id+".ct"))
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "reading metadata failed", http.StatusInternalServerError)
		return
	}

	f, err := os.Open(filepath.Join(h.dataDir, id))
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "reading file failed", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", string(ctBytes))
	io.Copy(w, f)
}

// Delete handles DELETE /files/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, err := os.Stat(filepath.Join(h.dataDir, id)); os.IsNotExist(err) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	os.Remove(filepath.Join(h.dataDir, id))
	os.Remove(filepath.Join(h.dataDir, id+".ct"))

	w.WriteHeader(http.StatusNoContent)
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
