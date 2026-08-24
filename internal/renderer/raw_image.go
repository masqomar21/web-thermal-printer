package renderer

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/masqomar21/antrean-ticket-printer/internal/printer"
)

// RenderRawImage decodes image byte stream (PNG/JPEG/GIF) and converts it to ESC/POS raster binary
func RenderRawImage(imgBytes []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	b := printer.NewESCPOSBuilder()
	b.AlignCenter().RasterImage(img).NewLine(3).CutPaper()
	return b.Bytes(), nil
}
