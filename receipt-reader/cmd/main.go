package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Joel-Haeberli/receipt-reader/internal/extractor"
	"github.com/Joel-Haeberli/receipt-reader/internal/handler"
)

func main() {
	lang := os.Getenv("TESSERACT_LANG")
	if lang == "" {
		lang = "eng"
	}

	fileServiceURL := os.Getenv("FILE_SERVICE_URL")
	if fileServiceURL == "" {
		fileServiceURL = "http://localhost:8083"
	}

	ext := extractor.NewTesseract(lang).
		WithStep(extractor.DownscaleStep).     // cap resolution at 3000px before any processing
		WithStep(extractor.RemoveAlphaStep).   // flatten transparency to white background
		WithStep(extractor.GrayscaleStep).     // colour → gray
		WithStep(extractor.ShadowRemovalStep). // normalise uneven illumination
		WithStep(extractor.GammaStep).         // auto brightness correction
		WithStep(extractor.OrientationStep).   // detect & correct 90°/180°/270° rotation via OSD
		WithStep(extractor.DeskewStep).        // straighten small skew — now before resize (cheaper)
		WithStep(extractor.ResizeStep).        // upscale to ≥400 DPI
		WithStep(extractor.UnsharpMaskStep).   // recover edge detail after resize
		WithStep(extractor.RemoveBordersStep). // crop uniform borders
		WithStep(extractor.MedianFilterStep).  // remove noise before CLAHE (noise skews tile histograms)
		WithStep(extractor.CLAHEStep).         // local contrast enhancement
		WithStep(extractor.SauvolaStep).       // adaptive binarisation
		WithStep(extractor.MorphCloseStep)     // fill gaps in character strokes
	h := handler.NewReceiptHandler(ext, fileServiceURL)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /receipts/extract", h.Extract)

	outerMux := http.NewServeMux()
	outerMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	auth := handler.NewAuthMiddleware("receipt_reader_read", "receipt_reader_write")
	outerMux.Handle("/", auth(mux))

	addr := ":8080"
	log.Printf("receipt-reader listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler.CORSMiddleware(handler.LoggingMiddleware(outerMux))))
}
