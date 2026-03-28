package model

import (
	"encoding/json"
	"io"
	"time"
)

// Account represents a financial account that receipts can be linked to.
type Account struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Shortcut  string    `json:"shortcut"`
	Budget    float64   `json:"budget"`
	ClassID   string    `json:"classId"`
	ValidFrom time.Time `json:"validFrom"`
	ValidTo   time.Time `json:"validTo"`

	// Transient — computed from linked receipt splits, not stored in the database.
	Spent   float64 `json:"spent"`
	Balance float64 `json:"balance"`
}

// EncodeAccount writes a as JSON to w.
func EncodeAccount(w io.Writer, a *Account) error {
	return json.NewEncoder(w).Encode(a)
}

// DecodeAccount reads a JSON-encoded Account from r.
func DecodeAccount(r io.Reader) (*Account, error) {
	var a Account
	if err := json.NewDecoder(r).Decode(&a); err != nil {
		return nil, err
	}
	return &a, nil
}
