package handler

import (
	_ "embed"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
	model "github.com/Joel-Haeberli/leanschool-model"
	"github.com/Joel-Haeberli/leanschool/internal/storage"
)

//go:embed receipts_template.xlsx
var receiptTemplateEmbedded []byte

// loadReceiptTemplate returns the template bytes. When RECEIPT_TEMPLATE_PATH is
// set it reads the file from disk; otherwise it falls back to the embedded copy.
func loadReceiptTemplate() ([]byte, error) {
	if path := os.Getenv("RECEIPT_TEMPLATE_PATH"); path != "" {
		return os.ReadFile(path)
	}
	return receiptTemplateEmbedded, nil
}

// fillPlaceholders replaces [PLACEHOLDER] tokens in all sheets of f.
// It searches every sheet for cells whose value contains a placeholder key
// and substitutes it with the corresponding value.
func fillPlaceholders(f *excelize.File, replacements map[string]string) {
	for _, sheet := range f.GetSheetList() {
		for placeholder, value := range replacements {
			cells, err := f.SearchSheet(sheet, regexp.QuoteMeta(placeholder), true)
			if err != nil {
				continue
			}
			for _, cell := range cells {
				current, _ := f.GetCellValue(sheet, cell)
				f.SetCellStr(sheet, cell, strings.ReplaceAll(current, placeholder, value))
			}
		}
	}
}

// ReceiptHandler handles CRUD HTTP requests for receipts.
type ReceiptHandler struct {
	store storage.Storage
}

func NewReceiptHandler(s storage.Storage) *ReceiptHandler {
	return &ReceiptHandler{store: s}
}

// RegisterRoutes registers all receipt routes on the given mux.
func (h *ReceiptHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /receipts", h.Create)
	mux.HandleFunc("GET /receipts", h.List)
	mux.HandleFunc("GET /receipts/{id}", h.Get)
	mux.HandleFunc("PUT /receipts/{id}", h.Update)
	mux.HandleFunc("DELETE /receipts/{id}", h.Delete)
	mux.HandleFunc("POST /receipts/export", h.Export)
	mux.HandleFunc("POST /receipts/submit", h.Submit)
}

// Create handles POST /receipts
func (h *ReceiptHandler) Create(w http.ResponseWriter, r *http.Request) {
	receipt, err := model.Decode(r.Body)
	if err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.Create(r.Context(), receipt); err != nil {
		http.Error(w, "create failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	model.Encode(w, receipt)
}

// Get handles GET /receipts/{id}
func (h *ReceiptHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	receipt, err := h.store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if receipt == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	model.Encode(w, receipt)
}

// List handles GET /receipts?owner=<id>
func (h *ReceiptHandler) List(w http.ResponseWriter, r *http.Request) {
	ownerID := r.URL.Query().Get("owner")

	receipts, err := h.store.List(r.Context(), ownerID)
	if err != nil {
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if receipts == nil {
		receipts = []*model.Receipt{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipts)
}

// Update handles PUT /receipts/{id}
func (h *ReceiptHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	receipt, err := model.Decode(r.Body)
	if err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	receipt.ID = id

	if err := h.store.Update(r.Context(), receipt); err != nil {
		if err.Error() == "receipt not found" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	model.Encode(w, receipt)
}

// Delete handles DELETE /receipts/{id}
func (h *ReceiptHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.store.Delete(r.Context(), id); err != nil {
		if err.Error() == "receipt not found" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Export handles POST /receipts/export
// Body: {"ids": ["id1", "id2", ...]}
// Returns an xlsx file populated from the receipts template.
func (h *ReceiptHandler) Export(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		http.Error(w, "invalid body: expected {\"ids\":[...]}", http.StatusBadRequest)
		return
	}

	// Fetch all receipts and index the requested ones.
	all, err := h.store.List(r.Context(), "")
	if err != nil {
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	idSet := make(map[string]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		idSet[id] = struct{}{}
	}
	var selected []*model.Receipt
	for _, rc := range all {
		if _, ok := idSet[rc.ID]; ok {
			selected = append(selected, rc)
		}
	}

	// Build account-shortcut lookup.
	accounts, err := h.store.ListAccounts(r.Context())
	if err != nil {
		http.Error(w, "list accounts failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shortcut := make(map[string]string, len(accounts))
	for _, a := range accounts {
		shortcut[a.ID] = a.Shortcut
	}

	// Resolve personal data from the authenticated user.
	var userName, userAddress, userIBAN string
	if claims := ClaimsFromContext(r.Context()); claims != nil {
		userName = claims.Name
		if userName == "" {
			userName = claims.PreferredUsername
		}
		if profile, err := h.store.GetUserProfile(r.Context(), claims.Sub); err == nil && profile != nil {
			userAddress = profile.Address
			userIBAN = profile.IBAN
		}
	}

	// Open template.
	tmplBytes, err := loadReceiptTemplate()
	if err != nil {
		http.Error(w, "load template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	f, err := excelize.OpenReader(bytes.NewReader(tmplBytes))
	if err != nil {
		http.Error(w, "open template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	const sheet = "PSW"
	const startRow = 10

	// Fill personal data by replacing [PLACEHOLDER] tokens across all sheets.
	fillPlaceholders(f, map[string]string{
		"[NAME] [PRENAME]":        userName,
		"[ADDRESS]":               userAddress,
		"[POSTAL_CODE] [LOCATION]": "",
		"[IBAN]":                  userIBAN,
	})

	rowOffset := 0
	for _, rc := range selected {
		dateStr := rc.Time.Format("02.01.2006")
		names := make([]string, 0, len(rc.Items))
		for _, it := range rc.Items {
			names = append(names, it.Name)
		}
		desc := strings.Join(names, ", ")

		if len(rc.Splits) > 0 {
			for _, sp := range rc.Splits {
				row := startRow + rowOffset
				cell := func(col string) string { return fmt.Sprintf("%s%d", col, row) }
				f.SetCellStr(sheet, cell("A"), dateStr)
				f.SetCellStr(sheet, cell("B"), desc)
				f.SetCellStr(sheet, cell("F"), shortcut[sp.AccountID])
				f.SetCellFloat(sheet, cell("G"), sp.Amount, 2, 64)
				rowOffset++
			}
		} else {
			row := startRow + rowOffset
			cell := func(col string) string { return fmt.Sprintf("%s%d", col, row) }
			f.SetCellStr(sheet, cell("A"), dateStr)
			f.SetCellStr(sheet, cell("B"), desc)
			f.SetCellStr(sheet, cell("F"), shortcut[rc.AccountID])
			f.SetCellFloat(sheet, cell("G"), rc.TotalPrice, 2, 64)
			rowOffset++
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		http.Error(w, "write xlsx: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="receipts.xlsx"`)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Write(buf.Bytes())
}

// Submit handles POST /receipts/submit
// Body: {"ids": ["id1", "id2", ...]}
// Sets the status of each receipt to "submitted".
func (h *ReceiptHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		http.Error(w, "invalid body: expected {\"ids\":[...]}", http.StatusBadRequest)
		return
	}
	if err := h.store.UpdateStatus(r.Context(), req.IDs, model.Submitted); err != nil {
		http.Error(w, "submit failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── AccountHandler ────────────────────────────────────────────────────────────

// AccountHandler handles CRUD HTTP requests for accounts.
type AccountHandler struct {
	store storage.Storage
}

func NewAccountHandler(s storage.Storage) *AccountHandler {
	return &AccountHandler{store: s}
}

// RegisterRoutes registers all account routes on the given mux.
func (h *AccountHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /accounts", h.Create)
	mux.HandleFunc("GET /accounts", h.List)
	mux.HandleFunc("GET /accounts/{id}", h.Get)
	mux.HandleFunc("PUT /accounts/{id}", h.Update)
	mux.HandleFunc("DELETE /accounts/{id}", h.Delete)
}

// Create handles POST /accounts
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	account, err := model.DecodeAccount(r.Body)
	if err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !hasRole(ClaimsFromContext(r.Context()), "school-management") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.store.CreateAccount(r.Context(), account); err != nil {
		http.Error(w, "create failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	model.EncodeAccount(w, account)
}

// List handles GET /accounts
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.store.ListAccounts(r.Context())
	if err != nil {
		http.Error(w, "list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if accounts == nil {
		accounts = []*model.Account{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// Get handles GET /accounts/{id}
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	account, err := h.store.GetAccount(r.Context(), id)
	if err != nil {
		http.Error(w, "get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if account == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	model.EncodeAccount(w, account)
}

// Update handles PUT /accounts/{id}
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	account, err := model.DecodeAccount(r.Body)
	if err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	account.ID = id

	if err := h.store.UpdateAccount(r.Context(), account); err != nil {
		if err.Error() == "account not found" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	model.EncodeAccount(w, account)
}

// Delete handles DELETE /accounts/{id}
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.store.DeleteAccount(r.Context(), id); err != nil {
		if err.Error() == "account not found" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
