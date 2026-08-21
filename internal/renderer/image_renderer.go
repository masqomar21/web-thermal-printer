package renderer

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/masqomar21/antrean-ticket-printer/internal/model"
	"github.com/masqomar21/antrean-ticket-printer/internal/printer"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// RenderImageTicket creates a graphic raster ticket image and returns ESC/POS command bytes
func RenderImageTicket(data model.TicketData, widthDots int) []byte {
	if widthDots <= 0 {
		widthDots = 576 // Default 80mm thermal printer (576 pixels wide)
	}

	heightDots := 320 // Height for ticket canvas
	img := image.NewRGBA(image.Rect(0, 0, widthDots, heightDots))

	// Fill background with white
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{white}, image.Point{}, draw.Src)

	// Draw text elements onto canvas
	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{black},
		Face: basicfont.Face7x13,
	}

	// 1. Header / Instansi
	headerText := "LOKET: " + data.Loket
	if data.Instansi != "" {
		headerText = data.Instansi + " - " + headerText
	}
	drawCenteredText(drawer, headerText, widthDots, 40, 2)

	// 2. Separator line
	drawLine(img, 20, 60, widthDots-20, 60, black)

	// 3. Nomor Antrean (Large font scale)
	drawCenteredText(drawer, data.NomorAntrean, widthDots, 150, 4)

	// 4. Separator line
	drawLine(img, 20, 220, widthDots-20, 220, black)

	// 5. Date & Time
	dateTimeText := "Tanggal: " + data.Tanggal + "  |  Jam: " + data.Jam
	drawCenteredText(drawer, dateTimeText, widthDots, 260, 1)

	if data.Layanan != "" {
		drawCenteredText(drawer, "Layanan: "+data.Layanan, widthDots, 285, 1)
	}

	// Build ESC/POS raster image bytes
	b := printer.NewESCPOSBuilder()
	b.AlignCenter().RasterImage(img).NewLine(3).CutPaper()

	return b.Bytes()
}

func drawCenteredText(drawer *font.Drawer, text string, totalWidth int, y int, scale int) {
	textWidth := font.MeasureString(drawer.Face, text).Ceil() * scale
	x := (totalWidth - textWidth) / 2
	if x < 10 {
		x = 10
	}

	if scale <= 1 {
		drawer.Dot = fixed.P(x, y)
		drawer.DrawString(text)
	} else {
		// Simple multi-pass scale draw for clear bold text
		for dy := 0; dy < scale; dy++ {
			for dx := 0; dx < scale; dx++ {
				drawer.Dot = fixed.P(x+dx, y+dy)
				drawer.DrawString(text)
			}
		}
	}
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for x := x1; x <= x2; x++ {
		for y := y1 - 1; y <= y1 + 1; y++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}
