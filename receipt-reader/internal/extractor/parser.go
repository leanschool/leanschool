package extractor

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	model "github.com/Joel-Haeberli/leanschool-model"
)

var (
	// rePrice matches prices like 1.50, 1,50, 1234.50, 1.234,50
	// Handles optional currency prefix/suffix (CHF, EUR, USD, €, $, Fr.)
	rePrice = regexp.MustCompile(`(?:CHF|EUR|USD|Fr\.?|€|\$)?\s*(\d{1,6})[.,](\d{2})\s*(?:CHF|EUR|USD|Fr\.?|€|\$)?`)

	// reTotal covers EN/DE/FR/IT total-line keywords.
	// t[o0]tal also catches common OCR misreads where 'o' is read as '0'.
	reTotal = regexp.MustCompile(`(?i)\b(t[o0]tal|totale|gesamt|summe|subtotal|amount\s+due|to\s+pay|zu\s+zahlen|zu\s+bezahlen|montant|montant\s+total|total\s+ttc|total\s+ht|betrag|endbetrag|rechnungsbetrag|da\s+pagare|importo|total\s+à\s+payer)\b`)

	// reTax covers EN/DE/FR/IT tax-line keywords
	reTax = regexp.MustCompile(`(?i)\b(tax|mwst|ust|vat|steuer|tva|iva|mva|mehrwertsteuer|incl\.?\s*tax|inkl\.?\s*mwst)\b`)

	// reDiscount: lines describing discounts/refunds — skip as items
	reDiscount = regexp.MustCompile(`(?i)\b(rabatt|discount|reduction|sconto|remise|abzug|gutschein|coupon|voucher)\b`)

	// rePayment: payment method lines — skip entirely
	rePayment = regexp.MustCompile(`(?i)\b(visa|mastercard|maestro|amex|twint|paypal|card|cash|kredit|debit|credit|ec[-\s]?karte|zahlung|payment|paiement|pagamento|bargeld)\b`)

	reDateTime = regexp.MustCompile(`(\d{1,2})[./-](\d{1,2})[./-](\d{2,4})\s+(\d{1,2}):(\d{2})`)
	reDate     = regexp.MustCompile(`(\d{1,2})[./-](\d{1,2})[./-](\d{2,4})`)
	reQtyXName = regexp.MustCompile(`^(\d+)\s*[xX*]\s+(.+)`)

	// reJunk: lines that are clearly non-data (only symbols, very short, bar codes)
	reJunk = regexp.MustCompile(`^[^a-zA-Z0-9]{0,2}$`)
)

// parseReceipt converts raw tesseract OCR text into a Receipt.
func parseReceipt(text string) *model.Receipt {
	receipt := &model.Receipt{Time: time.Now()}

	lines := strings.Split(text, "\n")
	log.Printf("[parser] processing %d lines", len(lines))

	// pendingName holds a descriptive line that had no price.  When the very
	// next item line has a very short name (< 3 chars — likely a stray code),
	// we prepend pendingName so multi-line item descriptions are joined.
	var pendingName string

	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || reJunk.MatchString(line) {
			continue
		}

		// Date/time detection
		if t := parseDateTime(line); !t.IsZero() {
			log.Printf("[parser] line %d: DATE      %q → %s", i, line, t.Format("2006-01-02 15:04"))
			receipt.Time = t
			pendingName = ""
			continue
		}

		// Total line — do not treat as item
		if reTotal.MatchString(line) {
			if p := lastPrice(line); p > 0 {
				log.Printf("[parser] line %d: TOTAL     %q → %.2f", i, line, p)
				receipt.TotalPrice = p
			} else {
				log.Printf("[parser] line %d: TOTAL     %q (no price found)", i, line)
			}
			pendingName = ""
			continue
		}

		// Tax line — do not treat as item
		if reTax.MatchString(line) {
			if p := lastPrice(line); p > 0 {
				log.Printf("[parser] line %d: TAX       %q → %.2f", i, line, p)
				receipt.Taxes = p
			} else {
				log.Printf("[parser] line %d: TAX       %q (no price found)", i, line)
			}
			pendingName = ""
			continue
		}

		// Discount/refund — skip
		if reDiscount.MatchString(line) {
			log.Printf("[parser] line %d: DISCOUNT  %q (skipped)", i, line)
			pendingName = ""
			continue
		}

		// Payment method line (VISA, CARD, CASH, etc.) — skip
		if rePayment.MatchString(line) {
			log.Printf("[parser] line %d: PAYMENT   %q (skipped)", i, line)
			pendingName = ""
			continue
		}

		// Item line heuristic: must contain a price-like pattern
		if item, ok := parseItemLine(line); ok {
			// If the parsed name is suspiciously short (stray code/digit), use
			// the pending descriptive line from above as the item name.
			if pendingName != "" && len(item.Name) < 3 {
				log.Printf("[parser] line %d: ITEM      joining pending %q + %q", i, pendingName, item.Name)
				item.Name = pendingName
			}
			pendingName = ""
			log.Printf("[parser] line %d: ITEM      %q → name=%q amount=%.0f price=%.2f", i, line, item.Name, item.Amount, item.Price)
			receipt.Items = append(receipt.Items, item)
		} else {
			// No price found — this might be the first line of a multi-line item name.
			if !isNumericOnly(line) && len(strings.Fields(line)) >= 1 {
				pendingName = line
				log.Printf("[parser] line %d: PENDING   %q", i, line)
			} else {
				log.Printf("[parser] line %d: SKIP      %q", i, line)
			}
		}
	}

	// If no explicit total was found, derive it from items.
	if receipt.TotalPrice == 0 && len(receipt.Items) > 0 {
		for _, it := range receipt.Items {
			receipt.TotalPrice += it.Amount * it.Price
		}
		log.Printf("[parser] total derived from items: %.2f", receipt.TotalPrice)
	}

	return receipt
}

// parseItemLine tries to parse a receipt line into a ReceiptItem.
// Supported formats:
//
//	"2x Milk           2.50"
//	"Milk              2.50"
//	"Milk     1    2.50"
//	"2.50 Milk"  (price-first format)
func parseItemLine(line string) (model.ReceiptItem, bool) {
	// Strip Swiss/French thousands separator before any price parsing.
	line = strings.ReplaceAll(line, "'", "")
	price := lastPrice(line)
	if price == 0 {
		return model.ReceiptItem{}, false
	}

	// Strip all price tokens to isolate the name.
	namePart := rePrice.ReplaceAllString(line, " ")
	namePart = strings.TrimSpace(namePart)
	namePart = strings.TrimRight(namePart, " \t-_.")

	amount := 1.0

	// Check for "2x Name", "2 x Name", "2 * Name" prefix.
	if m := reQtyXName.FindStringSubmatch(line); m != nil {
		if q, err := strconv.ParseFloat(m[1], 64); err == nil {
			amount = q
			namePart = rePrice.ReplaceAllString(m[2], "")
			namePart = strings.TrimSpace(namePart)
		}
	}

	// Strip trailing noise tokens: tax-code single chars (e.g. "e", "A") and
	// stray digits (line numbers, quantities) that appear after the price was removed.
	fields := strings.Fields(namePart)
	for len(fields) > 0 {
		last := fields[len(fields)-1]
		if len(last) == 1 || isNumericOnly(last) {
			fields = fields[:len(fields)-1]
		} else {
			break
		}
	}
	namePart = strings.Join(fields, " ")
	// Remove stray currency symbols left over.
	namePart = strings.Trim(namePart, "CHFEURUSDFr.€$ ")

	if namePart == "" || isNumericOnly(namePart) || len(namePart) < 2 {
		return model.ReceiptItem{}, false
	}

	return model.ReceiptItem{
		Name:   namePart,
		Amount: amount,
		Price:  price,
	}, true
}

// lastPrice returns the last decimal price found in s.
// Handles "1.50", "1,50", "CHF 1.50", "1.50 CHF", "1'234.50" (Swiss thousands separator).
func lastPrice(s string) float64 {
	// Strip Swiss/French thousands separator so "1'234.50" becomes "1234.50".
	s = strings.ReplaceAll(s, "'", "")
	matches := rePrice.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return 0
	}
	last := matches[len(matches)-1]
	intPart, _ := strconv.ParseFloat(last[1], 64)
	fracPart, _ := strconv.ParseFloat(last[2], 64)
	return intPart + fracPart/100.0
}

func parseDateTime(s string) time.Time {
	if m := reDateTime.FindStringSubmatch(s); m != nil {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		year := normalizeYear(m[3])
		hour, _ := strconv.Atoi(m[4])
		min, _ := strconv.Atoi(m[5])
		return time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC)
	}
	if m := reDate.FindStringSubmatch(s); m != nil {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		year := normalizeYear(m[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}
	return time.Time{}
}

func normalizeYear(s string) int {
	y, _ := strconv.Atoi(s)
	if y < 100 {
		y += 2000
	}
	return y
}

func isNumericOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) && r != '.' && r != ',' && r != ' ' {
			return false
		}
	}
	return true
}
