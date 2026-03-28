package extractor

import (
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ── Existing steps ────────────────────────────────────────────────────────────

// GrayscaleStep converts the image to grayscale.
// Grayscale images typically yield better OCR accuracy than colour originals.
var GrayscaleStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("grayscale: %w", err)
	}
	return encodeImage(toGray(src), "receipt-gray-*"+ext)
}

// RemoveAlphaStep composites the image onto a white background, eliminating any
// alpha channel. Transparent areas confuse Tesseract's binarisation heuristics.
var RemoveAlphaStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("remove-alpha: %w", err)
	}

	bounds := src.Bounds()
	flat := image.NewRGBA(bounds)
	draw.Draw(flat, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(flat, bounds, src, bounds.Min, draw.Over)

	return encodeImage(flat, "receipt-flat-*"+ext)
}

// DeskewStep straightens a slightly-rotated image using ImageMagick's -deskew.
// Requires ImageMagick to be installed; the step is skipped gracefully if not found.
// Run before ResizeStep so it operates on the smaller pre-resize image.
var DeskewStep OptimizationStep = func(ctx context.Context, imagePath string) (string, error) {
	if _, err := exec.LookPath("convert"); err != nil {
		return "", fmt.Errorf("deskew: ImageMagick 'convert' not found: %w", err)
	}

	ext := filepath.Ext(imagePath)
	out, err := os.CreateTemp("", "receipt-deskew-*"+ext)
	if err != nil {
		return "", fmt.Errorf("deskew: create temp: %w", err)
	}
	out.Close()

	// -deskew 40%: correct skew up to ~40% of the image width (generous threshold)
	// +repage: reset the virtual canvas origin after the transformation
	cmd := exec.CommandContext(ctx, "convert", imagePath, "-deskew", "40%", "+repage", out.Name())
	if b, err := cmd.CombinedOutput(); err != nil {
		os.Remove(out.Name())
		return "", fmt.Errorf("deskew: convert: %s: %w", strings.TrimSpace(string(b)), err)
	}
	return out.Name(), nil
}

// BinarizeStep applies Otsu's global thresholding (kept for reference).
// SauvolaStep is preferred for real-world receipt photos with uneven lighting.
var BinarizeStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("binarize: %w", err)
	}

	gray := toGray(src)
	bounds := gray.Bounds()
	threshold := otsuThreshold(gray)

	out := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if gray.GrayAt(x, y).Y >= threshold {
				out.SetGray(x, y, color.Gray{Y: 255})
			} else {
				out.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}
	return encodeImage(out, "receipt-binary-*"+ext)
}

// RemoveBordersStep crops uniform borders from all four sides of the image.
var RemoveBordersStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("remove-borders: %w", err)
	}

	b := src.Bounds()

	// Average background luminance sampled from the four corners.
	corners := [4][2]int{
		{b.Min.X, b.Min.Y},
		{b.Max.X - 1, b.Min.Y},
		{b.Min.X, b.Max.Y - 1},
		{b.Max.X - 1, b.Max.Y - 1},
	}
	var bgSum float64
	for _, c := range corners {
		r, g, bv, _ := src.At(c[0], c[1]).RGBA()
		bgSum += float64(r+g+bv) / 3
	}
	bg := bgSum / 4

	const threshold = 30 * 257.0

	isDiff := func(x, y int) bool {
		r, g, bv, _ := src.At(x, y).RGBA()
		return math.Abs(float64(r+g+bv)/3-bg) > threshold
	}

	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X-1, b.Min.Y-1

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if isDiff(x, y) {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if minX > maxX || minY > maxY {
		return imagePath, nil
	}

	const pad = 8
	minX = max(b.Min.X, minX-pad)
	minY = max(b.Min.Y, minY-pad)
	maxX = min(b.Max.X-1, maxX+pad)
	maxY = min(b.Max.Y-1, maxY+pad)

	type subImager interface {
		SubImage(image.Rectangle) image.Image
	}
	rect := image.Rect(minX, minY, maxX+1, maxY+1)
	var cropped image.Image
	if si, ok := src.(subImager); ok {
		cropped = si.SubImage(rect)
	} else {
		dst := image.NewNRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				dst.Set(x-minX, y-minY, src.At(x, y))
			}
		}
		cropped = dst
	}

	return encodeImage(cropped, "receipt-trimmed-*"+ext)
}

// ResizeStep upscales the image when its DPI metadata is below 400.
// Targeting 400 DPI (above Tesseract's 300 DPI minimum) noticeably improves
// small-font accuracy. Images without DPI metadata are left unchanged.
// The scale factor is capped at 8× to avoid runaway upscaling.
var ResizeStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	dpi := readDPI(imagePath)
	if dpi <= 0 || dpi >= 400 {
		return imagePath, nil
	}

	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("resize: %w", err)
	}

	scale := math.Min(400.0/dpi, 8.0)
	b := src.Bounds()
	newW := int(math.Round(float64(b.Dx()) * scale))
	newH := int(math.Round(float64(b.Dy()) * scale))

	return encodeImage(bilinearScale(src, newW, newH), "receipt-resized-*"+ext)
}

// ── New steps ─────────────────────────────────────────────────────────────────

// DownscaleStep caps the longer image dimension at 3000 px.
// Very large phone photos (12 MP+) would otherwise cause every subsequent step
// and Tesseract itself to be unnecessarily slow without improving accuracy.
// Images already within the limit are passed through unchanged.
var DownscaleStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("downscale: %w", err)
	}
	b := src.Bounds()
	W, H := b.Dx(), b.Dy()

	const maxDim = 3000
	longer := W
	if H > longer {
		longer = H
	}
	if longer <= maxDim {
		return imagePath, nil
	}

	scale := float64(maxDim) / float64(longer)
	newW := int(math.Round(float64(W) * scale))
	newH := int(math.Round(float64(H) * scale))
	log.Printf("[downscale] %dx%d → %dx%d (scale %.2f)", W, H, newW, newH, scale)
	return encodeImage(bilinearScale(src, newW, newH), "receipt-downscaled-*"+ext)
}

// OrientationStep uses Tesseract's Orientation and Script Detection (PSM 0)
// to detect if the image is rotated 90°, 180°, or 270° and corrects it.
// This handles photos taken with the phone held sideways or upside-down.
// Requires the osd.traineddata language file; the step is skipped gracefully
// if OSD data is unavailable or the confidence is too low.
var OrientationStep OptimizationStep = func(ctx context.Context, imagePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "tesseract", imagePath, "stdout", "--psm", "0", "--oem", "3")
	out, _ := cmd.CombinedOutput() // ignore exit error — OSD may warn but still emit orientation

	rotate := 0
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Rotate:") {
			fields := strings.Fields(strings.TrimPrefix(line, "Rotate:"))
			if len(fields) > 0 {
				rotate, _ = strconv.Atoi(fields[0])
			}
			break
		}
	}

	if rotate == 0 {
		return imagePath, nil
	}

	log.Printf("[orientation] rotating %d° CW to correct detected orientation", rotate)
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("orientation: %w", err)
	}
	return encodeImage(rotateImage(src, rotate), "receipt-oriented-*"+ext)
}

// ShadowRemovalStep normalises uneven illumination by dividing each grayscale
// pixel by a heavily blurred estimate of the local background brightness.
// This corrects shadows cast by a phone held over a receipt, making text
// contrast uniform across the whole image before binarisation.
var ShadowRemovalStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("shadow-removal: %w", err)
	}
	gray := toGray(src)
	b := gray.Bounds()

	// Large box blur estimates background illumination cheaply (O(W×H)).
	radius := max(30, min(b.Dx(), b.Dy())/10)
	bg := boxBlur(gray, radius)

	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			g := float64(gray.GrayAt(x, y).Y)
			bv := float64(bg.GrayAt(x, y).Y)
			var v float64
			if bv > 1 {
				v = g / bv * 255
			} else {
				v = g
			}
			out.SetGray(x, y, color.Gray{Y: clampUint8(v)})
		}
	}
	return encodeImage(out, "receipt-noshadow-*"+ext)
}

// GammaStep auto-detects over- or under-exposed images and applies a
// compensating power-law gamma correction. Images within a "normal" brightness
// range are passed through unchanged (no-op, returns the same path).
var GammaStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("gamma: %w", err)
	}
	gray := toGray(src)
	b := gray.Bounds()

	var sum int64
	n := int64(b.Dx() * b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			sum += int64(gray.GrayAt(x, y).Y)
		}
	}
	mean := float64(sum) / float64(n)

	var gamma float64
	switch {
	case mean > 180:
		gamma = 1.8 // overexposed — darken to recover contrast
	case mean < 70:
		gamma = 0.55 // underexposed — brighten
	default:
		return imagePath, nil // well-exposed — skip
	}

	var lut [256]uint8
	for i := range lut {
		lut[i] = clampUint8(math.Pow(float64(i)/255.0, gamma) * 255.0)
	}

	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			out.SetGray(x, y, color.Gray{Y: lut[gray.GrayAt(x, y).Y]})
		}
	}
	return encodeImage(out, "receipt-gamma-*"+ext)
}

// UnsharpMaskStep sharpens the image using the classic unsharp mask technique:
//
//	sharpened = original + amount × (original − blurred)
//
// Applied after ResizeStep to recover edge detail softened by bilinear upscaling.
var UnsharpMaskStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("unsharp-mask: %w", err)
	}
	gray := toGray(src)
	blurred := gaussianBlur(gray, 1.0) // small sigma — sharpen edges, not smooth
	b := gray.Bounds()

	const amount = 1.5
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			g := float64(gray.GrayAt(x, y).Y)
			bl := float64(blurred.GrayAt(x, y).Y)
			out.SetGray(x, y, color.Gray{Y: clampUint8(g + amount*(g-bl))})
		}
	}
	return encodeImage(out, "receipt-sharp-*"+ext)
}

// CLAHEStep applies Contrast Limited Adaptive Histogram Equalization.
// The image is divided into 64×64 tiles; each tile's histogram is independently
// equalised with bin clipping to avoid over-amplifying noise. The four nearest
// tile LUTs are bilinearly interpolated per pixel to avoid block artefacts.
var CLAHEStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("clahe: %w", err)
	}
	gray := toGray(src)
	b := gray.Bounds()
	W, H := b.Dx(), b.Dy()

	const tileSize = 64
	NX := (W + tileSize - 1) / tileSize
	NY := (H + tileSize - 1) / tileSize

	// Build one LUT per tile.
	luts := make([][256]uint8, NX*NY)
	for ty := 0; ty < NY; ty++ {
		for tx := 0; tx < NX; tx++ {
			x0 := b.Min.X + tx*tileSize
			x1 := min(x0+tileSize, b.Max.X)
			y0 := b.Min.Y + ty*tileSize
			y1 := min(y0+tileSize, b.Max.Y)

			var hist [256]int
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					hist[gray.GrayAt(x, y).Y]++
				}
			}
			nPixels := (x1 - x0) * (y1 - y0)
			clipLimit := max(1, 3*nPixels/256) // ~3× the average bin count
			luts[ty*NX+tx] = claheComputeLUT(hist, nPixels, clipLimit)
		}
	}

	// Apply with bilinear interpolation between the four nearest tile centres.
	out := image.NewGray(b)
	for py := b.Min.Y; py < b.Max.Y; py++ {
		for px := b.Min.X; px < b.Max.X; px++ {
			// Fractional tile index (tile centre sits at tile+0.5 in tile-space).
			ftx := float64(px-b.Min.X)/float64(tileSize) - 0.5
			fty := float64(py-b.Min.Y)/float64(tileSize) - 0.5

			tx0 := int(math.Floor(ftx))
			ty0 := int(math.Floor(fty))
			dx := ftx - float64(tx0)
			dy := fty - float64(ty0)

			tx0c := max(0, min(NX-1, tx0))
			tx1c := max(0, min(NX-1, tx0+1))
			ty0c := max(0, min(NY-1, ty0))
			ty1c := max(0, min(NY-1, ty0+1))

			val := gray.GrayAt(px, py).Y
			v00 := float64(luts[ty0c*NX+tx0c][val])
			v10 := float64(luts[ty0c*NX+tx1c][val])
			v01 := float64(luts[ty1c*NX+tx0c][val])
			v11 := float64(luts[ty1c*NX+tx1c][val])

			v := v00*(1-dx)*(1-dy) + v10*dx*(1-dy) + v01*(1-dx)*dy + v11*dx*dy
			out.SetGray(px, py, color.Gray{Y: clampUint8(v)})
		}
	}
	return encodeImage(out, "receipt-clahe-*"+ext)
}

// MedianFilterStep applies a 3×3 median filter to remove salt-and-pepper noise
// from camera grain or JPEG compression artefacts without blurring text edges.
// Placed before CLAHEStep so noise does not skew the per-tile histograms.
var MedianFilterStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("median-filter: %w", err)
	}
	gray := toGray(src)
	b := gray.Bounds()
	out := image.NewGray(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var neighbors [9]uint8
			k := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx := max(b.Min.X, min(b.Max.X-1, x+dx))
					ny := max(b.Min.Y, min(b.Max.Y-1, y+dy))
					neighbors[k] = gray.GrayAt(nx, ny).Y
					k++
				}
			}
			// Insertion sort on 9 elements — faster than sort.Slice for tiny arrays.
			for i := 1; i < 9; i++ {
				v := neighbors[i]
				j := i - 1
				for j >= 0 && neighbors[j] > v {
					neighbors[j+1] = neighbors[j]
					j--
				}
				neighbors[j+1] = v
			}
			out.SetGray(x, y, color.Gray{Y: neighbors[4]}) // median = index 4 of 9
		}
	}
	return encodeImage(out, "receipt-median-*"+ext)
}

// SauvolaStep applies Sauvola adaptive binarisation.
// Unlike global Otsu thresholding, it computes a per-pixel threshold from the
// local mean and standard deviation in a 31×31 window:
//
//	T(x,y) = mean × (1 + k × (stddev/R − 1))   k=0.2, R=128
//
// This makes it robust against uneven illumination — the most common failure
// mode for receipt photos taken on a smartphone.
// Summed-area tables give O(W×H) complexity regardless of window size.
var SauvolaStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("sauvola: %w", err)
	}
	gray := toGray(src)
	b := gray.Bounds()
	W, H := b.Dx(), b.Dy()

	// Build summed-area tables for mean and variance.
	sat1 := make([]int64, W*H) // sum of pixel values
	sat2 := make([]int64, W*H) // sum of squared pixel values
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			v := int64(gray.GrayAt(b.Min.X+x, b.Min.Y+y).Y)
			sat1[y*W+x] = v
			sat2[y*W+x] = v * v
			if x > 0 {
				sat1[y*W+x] += sat1[y*W+x-1]
				sat2[y*W+x] += sat2[y*W+x-1]
			}
			if y > 0 {
				sat1[y*W+x] += sat1[(y-1)*W+x]
				sat2[y*W+x] += sat2[(y-1)*W+x]
			}
			if x > 0 && y > 0 {
				sat1[y*W+x] -= sat1[(y-1)*W+x-1]
				sat2[y*W+x] -= sat2[(y-1)*W+x-1]
			}
		}
	}

	satQuery := func(sat []int64, x0, y0, x1, y1 int) int64 {
		s := sat[y1*W+x1]
		if x0 > 0 {
			s -= sat[y1*W+x0-1]
		}
		if y0 > 0 {
			s -= sat[(y0-1)*W+x1]
		}
		if x0 > 0 && y0 > 0 {
			s += sat[(y0-1)*W+x0-1]
		}
		return s
	}

	const (
		halfW = 15  // half-width of the 31×31 window
		k     = 0.2 // sensitivity to local standard deviation
		R     = 128.0
	)

	out := image.NewGray(b)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			x0 := max(0, x-halfW)
			x1 := min(W-1, x+halfW)
			y0 := max(0, y-halfW)
			y1 := min(H-1, y+halfW)
			n := int64((x1 - x0 + 1) * (y1 - y0 + 1))

			sum1 := satQuery(sat1, x0, y0, x1, y1)
			sum2 := satQuery(sat2, x0, y0, x1, y1)

			mean := float64(sum1) / float64(n)
			variance := float64(sum2)/float64(n) - mean*mean
			if variance < 0 {
				variance = 0
			}
			stddev := math.Sqrt(variance)

			threshold := mean * (1 + k*(stddev/R-1))
			pv := float64(gray.GrayAt(b.Min.X+x, b.Min.Y+y).Y)
			if pv < threshold {
				out.SetGray(b.Min.X+x, b.Min.Y+y, color.Gray{Y: 0})   // foreground (text)
			} else {
				out.SetGray(b.Min.X+x, b.Min.Y+y, color.Gray{Y: 255}) // background
			}
		}
	}
	return encodeImage(out, "receipt-sauvola-*"+ext)
}

// MorphCloseStep applies morphological closing (3×3 dilation then erosion) to
// the binarised image. This fills small gaps in character strokes that result
// from thin thermal print or slight binarisation noise, connecting broken
// letter parts that Tesseract's segmenter might otherwise split.
var MorphCloseStep OptimizationStep = func(_ context.Context, imagePath string) (string, error) {
	src, ext, err := decodeImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("morph-close: %w", err)
	}
	gray := toGray(src)
	b := gray.Bounds()

	// Dilation: a pixel becomes black (0) if any 3×3 neighbour is black.
	dilated := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			minVal := uint8(255)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx := max(b.Min.X, min(b.Max.X-1, x+dx))
					ny := max(b.Min.Y, min(b.Max.Y-1, y+dy))
					if v := gray.GrayAt(nx, ny).Y; v < minVal {
						minVal = v
					}
				}
			}
			dilated.SetGray(x, y, color.Gray{Y: minVal})
		}
	}

	// Erosion: a pixel stays black (0) only if all 3×3 neighbours are black.
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			maxVal := uint8(0)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx := max(b.Min.X, min(b.Max.X-1, x+dx))
					ny := max(b.Min.Y, min(b.Max.Y-1, y+dy))
					if v := dilated.GrayAt(nx, ny).Y; v > maxVal {
						maxVal = v
					}
				}
			}
			out.SetGray(x, y, color.Gray{Y: maxVal})
		}
	}
	return encodeImage(out, "receipt-morph-*"+ext)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// rotateImage rotates src clockwise by degrees (90, 180, or 270).
// For 90°/270° the output dimensions are swapped (H×W). Any other value
// returns the original image unchanged.
func rotateImage(src image.Image, degrees int) image.Image {
	b := src.Bounds()
	W, H := b.Dx(), b.Dy()

	switch degrees % 360 {
	case 90:
		// 90° CW: original (x,y) → new (H-1-y, x); new dims H×W
		dst := image.NewNRGBA(image.Rect(0, 0, H, W))
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				dst.Set(H-1-y, x, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 180:
		// 180°: original (x,y) → new (W-1-x, H-1-y); same dims
		dst := image.NewNRGBA(image.Rect(0, 0, W, H))
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				dst.Set(W-1-x, H-1-y, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 270:
		// 270° CW (= 90° CCW): original (x,y) → new (y, W-1-x); new dims H×W
		dst := image.NewNRGBA(image.Rect(0, 0, H, W))
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				dst.Set(y, W-1-x, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	}
	return src
}

// toGray converts any image to *image.Gray.
func toGray(src image.Image) *image.Gray {
	b := src.Bounds()
	gray := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			gray.Set(x, y, color.GrayModel.Convert(src.At(x, y)))
		}
	}
	return gray
}

// clampUint8 clamps v to [0, 255] and converts to uint8.
func clampUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// boxBlur blurs src using a (2*radius+1)² box kernel.
// Uses a summed-area table so complexity is O(W×H) regardless of radius.
func boxBlur(src *image.Gray, radius int) *image.Gray {
	b := src.Bounds()
	W, H := b.Dx(), b.Dy()

	sat := make([]int64, W*H)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			v := int64(src.GrayAt(b.Min.X+x, b.Min.Y+y).Y)
			if x > 0 {
				v += sat[y*W+x-1]
			}
			if y > 0 {
				v += sat[(y-1)*W+x]
			}
			if x > 0 && y > 0 {
				v -= sat[(y-1)*W+x-1]
			}
			sat[y*W+x] = v
		}
	}

	query := func(x0, y0, x1, y1 int) int64 {
		s := sat[y1*W+x1]
		if x0 > 0 {
			s -= sat[y1*W+x0-1]
		}
		if y0 > 0 {
			s -= sat[(y0-1)*W+x1]
		}
		if x0 > 0 && y0 > 0 {
			s += sat[(y0-1)*W+x0-1]
		}
		return s
	}

	out := image.NewGray(b)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			x0 := max(0, x-radius)
			x1 := min(W-1, x+radius)
			y0 := max(0, y-radius)
			y1 := min(H-1, y+radius)
			n := int64((x1 - x0 + 1) * (y1 - y0 + 1))
			out.SetGray(b.Min.X+x, b.Min.Y+y, color.Gray{Y: uint8(query(x0, y0, x1, y1) / n)})
		}
	}
	return out
}

// gaussianBlur applies a separable Gaussian blur with the given sigma.
// Kernel radius = ceil(3σ); elements beyond 3σ contribute < 0.1%.
func gaussianBlur(src *image.Gray, sigma float64) *image.Gray {
	radius := int(math.Ceil(3 * sigma))
	size := 2*radius + 1
	kernel := make([]float64, size)
	sum := 0.0
	for i := range kernel {
		x := float64(i - radius)
		kernel[i] = math.Exp(-x * x / (2 * sigma * sigma))
		sum += kernel[i]
	}
	for i := range kernel {
		kernel[i] /= sum
	}

	b := src.Bounds()

	// Horizontal pass.
	tmp := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var acc float64
			for ki, kv := range kernel {
				xi := max(b.Min.X, min(b.Max.X-1, x+ki-radius))
				acc += float64(src.GrayAt(xi, y).Y) * kv
			}
			tmp.SetGray(x, y, color.Gray{Y: clampUint8(acc)})
		}
	}

	// Vertical pass.
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var acc float64
			for ki, kv := range kernel {
				yi := max(b.Min.Y, min(b.Max.Y-1, y+ki-radius))
				acc += float64(tmp.GrayAt(x, yi).Y) * kv
			}
			out.SetGray(x, y, color.Gray{Y: clampUint8(acc)})
		}
	}
	return out
}

// claheComputeLUT clips a histogram at clipLimit, redistributes the excess
// uniformly, then returns a 256-entry mapping from old intensity to new.
func claheComputeLUT(hist [256]int, nPixels, clipLimit int) [256]uint8 {
	excess := 0
	for i := range hist {
		if hist[i] > clipLimit {
			excess += hist[i] - clipLimit
			hist[i] = clipLimit
		}
	}
	perBin := excess / 256
	rem := excess % 256
	for i := range hist {
		hist[i] += perBin
		if i < rem {
			hist[i]++
		}
	}
	var lut [256]uint8
	cdf := 0
	scale := 255.0 / float64(nPixels)
	for i := range hist {
		cdf += hist[i]
		lut[i] = clampUint8(float64(cdf) * scale)
	}
	return lut
}

// otsuThreshold computes the optimal global binarisation threshold using
// Otsu's method (maximises inter-class variance).
func otsuThreshold(img *image.Gray) uint8 {
	bounds := img.Bounds()
	total := (bounds.Max.X - bounds.Min.X) * (bounds.Max.Y - bounds.Min.Y)

	var hist [256]int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			hist[img.GrayAt(x, y).Y]++
		}
	}

	var sumAll float64
	for i, c := range hist {
		sumAll += float64(i) * float64(c)
	}

	var sumB float64
	var wB int
	var best float64
	var threshold uint8

	for i, c := range hist {
		wB += c
		if wB == 0 {
			continue
		}
		wF := total - wB
		if wF == 0 {
			break
		}
		sumB += float64(i) * float64(c)
		mB := sumB / float64(wB)
		mF := (sumAll - sumB) / float64(wF)
		v := float64(wB) * float64(wF) * (mB - mF) * (mB - mF)
		if v > best {
			best = v
			threshold = uint8(i)
		}
	}
	return threshold
}

// readDPI returns the X resolution in DPI from JPEG or PNG metadata.
// Returns 0 when the format is unsupported or the metadata is absent/unit-less.
func readDPI(imagePath string) float64 {
	f, err := os.Open(imagePath)
	if err != nil {
		return 0
	}
	defer f.Close()

	switch strings.ToLower(filepath.Ext(imagePath)) {
	case ".jpg", ".jpeg":
		dpi, _ := readJPEGDPI(f)
		return dpi
	case ".png":
		dpi, _ := readPNGDPI(f)
		return dpi
	default:
		return 0
	}
}

// readJPEGDPI reads the X resolution from the JFIF APP0 segment.
func readJPEGDPI(r io.Reader) (float64, error) {
	var h [2]byte
	if _, err := io.ReadFull(r, h[:]); err != nil {
		return 0, err
	}
	if h[0] != 0xFF || h[1] != 0xD8 {
		return 0, nil
	}
	var buf [2]byte
	for {
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return 0, nil
		}
		if buf[0] != 0xFF {
			return 0, nil
		}
		marker := buf[1]
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return 0, nil
		}
		length := int(binary.BigEndian.Uint16(buf[:])) - 2
		if length < 0 {
			return 0, nil
		}
		data := make([]byte, length)
		if _, err := io.ReadFull(r, data); err != nil {
			return 0, nil
		}
		if marker == 0xE0 && length >= 12 && string(data[:4]) == "JFIF" {
			xDensity := float64(binary.BigEndian.Uint16(data[6:8]))
			switch data[5] {
			case 1:
				return xDensity, nil
			case 2:
				return xDensity * 2.54, nil
			}
			return 0, nil
		}
		if marker == 0xDA {
			return 0, nil
		}
	}
}

// readPNGDPI reads the X resolution from the PNG pHYs chunk.
func readPNGDPI(r io.Reader) (float64, error) {
	if _, err := io.ReadFull(r, make([]byte, 8)); err != nil {
		return 0, err
	}
	var lenBuf, typeBuf [4]byte
	for {
		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			return 0, nil
		}
		chunkLen := int(binary.BigEndian.Uint32(lenBuf[:]))
		if _, err := io.ReadFull(r, typeBuf[:]); err != nil {
			return 0, nil
		}
		data := make([]byte, chunkLen)
		if _, err := io.ReadFull(r, data); err != nil {
			return 0, nil
		}
		if _, err := io.ReadFull(r, make([]byte, 4)); err != nil {
			return 0, nil
		}
		switch string(typeBuf[:]) {
		case "pHYs":
			if chunkLen != 9 {
				return 0, nil
			}
			xPPU := float64(binary.BigEndian.Uint32(data[0:4]))
			if data[8] == 1 {
				return xPPU * 0.0254, nil
			}
			return 0, nil
		case "IDAT", "IEND":
			return 0, nil
		}
	}
}

// bilinearScale resizes src to newW×newH using bilinear interpolation.
func bilinearScale(src image.Image, newW, newH int) image.Image {
	b := src.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, newW, newH))

	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			fx := (float64(x)+0.5)*float64(srcW)/float64(newW) - 0.5
			fy := (float64(y)+0.5)*float64(srcH)/float64(newH) - 0.5

			x0, y0 := int(fx), int(fy)
			x1, y1 := x0+1, y0+1
			dx, dy := fx-float64(x0), fy-float64(y0)

			x0 = max(b.Min.X, min(b.Max.X-1, x0))
			y0 = max(b.Min.Y, min(b.Max.Y-1, y0))
			x1 = max(b.Min.X, min(b.Max.X-1, x1))
			y1 = max(b.Min.Y, min(b.Max.Y-1, y1))

			r00, g00, b00, a00 := src.At(x0, y0).RGBA()
			r10, g10, b10, a10 := src.At(x1, y0).RGBA()
			r01, g01, b01, a01 := src.At(x0, y1).RGBA()
			r11, g11, b11, a11 := src.At(x1, y1).RGBA()

			lerp := func(c00, c10, c01, c11 uint32) uint8 {
				top := float64(c00)*(1-dx) + float64(c10)*dx
				bot := float64(c01)*(1-dx) + float64(c11)*dx
				return uint8((top*(1-dy)+bot*dy) / 256)
			}
			dst.SetNRGBA(x, y, color.NRGBA{
				R: lerp(r00, r10, r01, r11),
				G: lerp(g00, g10, g01, g11),
				B: lerp(b00, b10, b01, b11),
				A: lerp(a00, a10, a01, a11),
			})
		}
	}
	return dst
}

// decodeImage opens and decodes an image file, returning the image and its extension.
func decodeImage(imagePath string) (image.Image, string, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return nil, "", fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, "", fmt.Errorf("decode: %w", err)
	}
	return img, filepath.Ext(imagePath), nil
}

// encodeImage writes img to a new temp file named by the given pattern and
// returns its path. The caller is responsible for removing the file.
func encodeImage(img image.Image, pattern string) (string, error) {
	out, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	defer out.Close()

	ext := strings.ToLower(filepath.Ext(out.Name()))
	switch ext {
	case ".png":
		err = png.Encode(out, img)
	default:
		err = jpeg.Encode(out, img, &jpeg.Options{Quality: 95})
	}
	if err != nil {
		os.Remove(out.Name())
		return "", fmt.Errorf("encode: %w", err)
	}
	return out.Name(), nil
}
