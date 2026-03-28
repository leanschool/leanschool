package model

import (
	"encoding/json"
	"io"
)

// Decode reads a Receipt from r (JSON).
func Decode(r io.Reader) (*Receipt, error) {
	var receipt Receipt
	if err := json.NewDecoder(r).Decode(&receipt); err != nil {
		return nil, err
	}
	return &receipt, nil
}

// Encode writes receipt to w as JSON.
func Encode(w io.Writer, receipt *Receipt) error {
	return json.NewEncoder(w).Encode(receipt)
}
