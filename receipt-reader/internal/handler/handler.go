package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/receipt-reader/internal/extractor"
)

// ReceiptHandler handles receipt-related HTTP requests.
type ReceiptHandler struct {
	extractor      extractor.Extractor
	fileServiceURL string
}

func NewReceiptHandler(e extractor.Extractor, fileServiceURL string) *ReceiptHandler {
	return &ReceiptHandler{extractor: e, fileServiceURL: fileServiceURL}
}

// Extract handles POST /receipts/extract.
// Expects a JSON body: {"fileId": "...", "owner": "..."}.
// Fetches the image from the file-service, runs OCR, and returns the receipt.
func (h *ReceiptHandler) Extract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FileID string `json:"fileId"`
		Owner  string `json:"owner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FileID == "" {
		http.Error(w, `invalid body: expected {"fileId":"...","owner":"..."}`, http.StatusBadRequest)
		return
	}

	log.Printf("[handler] extract request: fileId=%q owner=%q", req.FileID, req.Owner)

	fileReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, fmt.Sprintf("%s/files/%s", h.fileServiceURL, req.FileID), nil)
	if err != nil {
		http.Error(w, "creating file request failed", http.StatusInternalServerError)
		return
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		fileReq.Header.Set("Authorization", auth)
	}
	imageResp, err := http.DefaultClient.Do(fileReq)
	if err != nil {
		log.Printf("[handler] fetching image from file-service: %v", err)
		http.Error(w, "fetching image failed", http.StatusBadGateway)
		return
	}
	defer imageResp.Body.Close()

	if imageResp.StatusCode != http.StatusOK {
		log.Printf("[handler] file-service returned %d for fileId=%s", imageResp.StatusCode, req.FileID)
		http.Error(w, "image not found in file-service", http.StatusNotFound)
		return
	}

	mediaType := imageResp.Header.Get("Content-Type")
	if mediaType == "" {
		mediaType = "image/jpeg"
	}

	receipt, err := h.extractor.Extract(r.Context(), imageResp.Body, mediaType)
	if err != nil {
		log.Printf("[handler] extraction failed: %v", err)
		http.Error(w, "extraction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	receipt.ReceiptOwnerID = req.Owner
	receipt.FileID = req.FileID

	log.Printf("[handler] extraction done: fileId=%s items=%d total=%.2f", req.FileID, len(receipt.Items), receipt.TotalPrice)

	w.Header().Set("Content-Type", "application/json")
	if err := model.Encode(w, receipt); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
