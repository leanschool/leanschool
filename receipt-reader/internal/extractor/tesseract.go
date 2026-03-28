package extractor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	model "github.com/Joel-Haeberli/leanschool-model"
)

// OptimizationStep is a pre-processing function applied to the image before OCR.
// It receives the path of the current image file, performs its transformation,
// writes the result to a new temp file, and returns its path.
// The caller is responsible for cleaning up the returned file.
type OptimizationStep func(ctx context.Context, imagePath string) (string, error)

// Tesseract extracts receipt data using the local tesseract binary.
type Tesseract struct {
	Lang  string
	steps []OptimizationStep
}

func NewTesseract(lang string) *Tesseract {
	if lang == "" {
		lang = "eng"
	}
	return &Tesseract{Lang: lang}
}

// WithStep appends an optimization step to the pre-processing pipeline.
// Steps are executed in the order they are added.
func (t *Tesseract) WithStep(step OptimizationStep) *Tesseract {
	t.steps = append(t.steps, step)
	return t
}

// Extract implements Extractor.
func (t *Tesseract) Extract(ctx context.Context, image io.Reader, mediaType string) (*model.Receipt, error) {
	ext := extensionFromMediaType(mediaType)
	tmp, err := os.CreateTemp("", "receipt-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	n, err := io.Copy(tmp, image)
	tmp.Close()
	if err != nil {
		return nil, fmt.Errorf("writing temp file: %w", err)
	}
	log.Printf("[tesseract] image saved: file=%s size=%d bytes mediaType=%s lang=%s", tmp.Name(), n, mediaType, t.Lang)

	// Run the pre-processing pipeline. Each step produces a new temp file;
	// we defer its removal immediately after creation so nothing leaks.
	ocrPath := tmp.Name()
	for i, step := range t.steps {
		out, err := step(ctx, ocrPath)
		if err != nil {
			log.Printf("[tesseract] optimization step %d failed: %v — skipping step", i, err)
			continue
		}
		defer os.Remove(out)
		log.Printf("[tesseract] optimization step %d applied: %s → %s", i, ocrPath, out)
		ocrPath = out
	}

	// Try PSM 4 (single column, variable sizes — best for most receipts) and
	// PSM 6 (uniform block); pick whichever produces the richer receipt.
	tsvPSM4, err := runTesseract(ctx, ocrPath, t.Lang, "4")
	if err != nil {
		log.Printf("[tesseract] psm=4 failed: %v — falling back to psm=6", err)
		tsvPSM4 = ""
	}
	tsvPSM6, err := runTesseract(ctx, ocrPath, t.Lang, "6")
	if err != nil {
		if tsvPSM4 == "" {
			return nil, fmt.Errorf("tesseract failed: %w", err)
		}
		log.Printf("[tesseract] psm=6 failed: %v — using psm=4 result only", err)
		tsvPSM6 = ""
	}

	textPSM4 := tsvToText(tsvPSM4, 30)
	textPSM6 := tsvToText(tsvPSM6, 30)

	log.Printf("[tesseract] psm=4 filtered text:\n---\n%s\n---", textPSM4)
	log.Printf("[tesseract] psm=6 filtered text:\n---\n%s\n---", textPSM6)

	receiptPSM4 := parseReceipt(textPSM4)
	receiptPSM6 := parseReceipt(textPSM6)

	score4 := receiptScore(receiptPSM4)
	score6 := receiptScore(receiptPSM6)

	receipt := receiptPSM4
	if score6 > score4 {
		receipt = receiptPSM6
		log.Printf("[tesseract] using psm=6 (score=%d > psm=4 score=%d)", score6, score4)
	} else {
		log.Printf("[tesseract] using psm=4 (score=%d >= psm=6 score=%d)", score4, score6)
	}

	log.Printf("[tesseract] parse result: items=%d total=%.2f taxes=%.2f date=%s",
		len(receipt.Items), receipt.TotalPrice, receipt.Taxes, receipt.Time.Format("2006-01-02"))

	return receipt, nil
}

// receiptScore rates how complete a parsed receipt is.
// Items are the primary signal; total and tax presence are bonuses.
func receiptScore(r *model.Receipt) int {
	s := len(r.Items) * 2
	if r.TotalPrice > 0 {
		s += 3
	}
	if r.Taxes > 0 {
		s += 2
	}
	return s
}

// runTesseract runs tesseract and returns raw TSV output.
// TSV output gives per-word confidence scores we use to filter noise.
// Extra dawg flags disable dictionary correction for prices, codes, and abbreviations.
func runTesseract(ctx context.Context, path, lang, psm string) (string, error) {
	cmd := exec.CommandContext(ctx, "tesseract", path, "stdout",
		"-l", lang,
		"--oem", "3",
		"--psm", psm,
		"--dpi", "400",
		"-c", "load_system_dawg=0",
		"-c", "load_freq_dawg=0",
		"-c", "load_punc_dawg=0",
		"-c", "load_number_dawg=0",
		"-c", "preserve_interword_spaces=1",
		"tsv",
	)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			if stderr != "" {
				log.Printf("[tesseract] psm=%s stderr: %s", psm, stderr)
			}
			return "", fmt.Errorf("psm=%s exit: %s", psm, stderr)
		}
		return "", fmt.Errorf("psm=%s: %w", psm, err)
	}
	return string(out), nil
}

// tsvToText converts Tesseract TSV output to plain text, discarding words
// whose confidence is below minConf. Words are grouped back into lines; a
// blank line is inserted between paragraphs.
//
// TSV columns (tab-separated, header on first line):
//
//	level page_num block_num par_num line_num word_num left top width height conf text
func tsvToText(tsv string, minConf float64) string {
	type lineKey struct{ block, par, line int }

	// ordered list of line keys so we can reconstruct top-to-bottom order
	var order []lineKey
	seen := map[lineKey]bool{}
	words := map[lineKey][]string{}

	lines := strings.Split(tsv, "\n")
	for _, raw := range lines {
		fields := strings.Split(raw, "\t")
		if len(fields) < 12 {
			continue
		}
		level, err := strconv.Atoi(fields[0])
		if err != nil || level != 5 {
			// level 5 = word; lower levels are page/block/par/line containers
			continue
		}
		conf, err := strconv.ParseFloat(fields[10], 64)
		if err != nil || conf < minConf {
			continue
		}
		word := strings.TrimSpace(fields[11])
		if word == "" {
			continue
		}
		block, _ := strconv.Atoi(fields[2])
		par, _ := strconv.Atoi(fields[3])
		line, _ := strconv.Atoi(fields[4])

		key := lineKey{block, par, line}
		if !seen[key] {
			seen[key] = true
			order = append(order, key)
		}
		words[key] = append(words[key], word)
	}

	var sb strings.Builder
	prevKey := lineKey{-1, -1, -1}
	for _, key := range order {
		if prevKey.block != -1 && (key.block != prevKey.block || key.par != prevKey.par) {
			sb.WriteByte('\n') // blank line between paragraphs/blocks
		}
		sb.WriteString(strings.Join(words[key], " "))
		sb.WriteByte('\n')
		prevKey = key
	}
	return sb.String()
}

func extensionFromMediaType(mt string) string {
	switch {
	case strings.Contains(mt, "png"):
		return ".png"
	case strings.Contains(mt, "gif"):
		return ".gif"
	case strings.Contains(mt, "webp"):
		return ".webp"
	case strings.Contains(mt, "tiff"):
		return ".tiff"
	default:
		return ".jpg"
	}
}
