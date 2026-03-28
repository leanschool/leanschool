package extractor

import (
	"context"
	"io"

	model "github.com/Joel-Haeberli/leanschool-model"
)

// Extractor extracts receipt data from an image.
type Extractor interface {
	Extract(ctx context.Context, image io.Reader, mediaType string) (*model.Receipt, error)
}
