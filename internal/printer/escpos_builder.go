package printer

import (
	"bytes"
	"image"
	"image/color"
)

// ESCPOSBuilder builds ESC/POS command byte sequences
type ESCPOSBuilder struct {
	buf bytes.Buffer
}

func NewESCPOSBuilder() *ESCPOSBuilder {
	b := &ESCPOSBuilder{}
	b.Init()
	return b
}

func (b *ESCPOSBuilder) Init() *ESCPOSBuilder {
	b.buf.Write([]byte{0x1B, 0x40}) // ESC @ (Initialize printer)
	return b
}

func (b *ESCPOSBuilder) AlignCenter() *ESCPOSBuilder {
	b.buf.Write([]byte{0x1B, 0x61, 0x01}) // ESC a 1
	return b
}

func (b *ESCPOSBuilder) AlignLeft() *ESCPOSBuilder {
	b.buf.Write([]byte{0x1B, 0x61, 0x00}) // ESC a 0
	return b
}

func (b *ESCPOSBuilder) AlignRight() *ESCPOSBuilder {
	b.buf.Write([]byte{0x1B, 0x61, 0x02}) // ESC a 2
	return b
}

func (b *ESCPOSBuilder) SetFontSize(widthMulti, heightMulti uint8) *ESCPOSBuilder {
	if widthMulti > 7 {
		widthMulti = 7
	}
	if heightMulti > 7 {
		heightMulti = 7
	}
	n := (widthMulti << 4) | heightMulti
	b.buf.Write([]byte{0x1D, 0x21, n}) // GS ! n
	return b
}

func (b *ESCPOSBuilder) SetBold(enable bool) *ESCPOSBuilder {
	if enable {
		b.buf.Write([]byte{0x1B, 0x45, 0x01}) // ESC E 1
	} else {
		b.buf.Write([]byte{0x1B, 0x45, 0x00}) // ESC E 0
	}
	return b
}

func (b *ESCPOSBuilder) Text(s string) *ESCPOSBuilder {
	b.buf.WriteString(s)
	return b
}

func (b *ESCPOSBuilder) TextLn(s string) *ESCPOSBuilder {
	b.buf.WriteString(s + "\n")
	return b
}

func (b *ESCPOSBuilder) NewLine(lines int) *ESCPOSBuilder {
	for i := 0; i < lines; i++ {
		b.buf.WriteByte('\n')
	}
	return b
}

func (b *ESCPOSBuilder) CutPaper() *ESCPOSBuilder {
	b.buf.Write([]byte{0x1D, 0x56, 0x41, 0x00}) // GS V 65 0 (Full cut with feed)
	return b
}

// RasterImage converts a Go image.Image to ESC/POS GS v 0 raster command format
func (b *ESCPOSBuilder) RasterImage(img image.Image) *ESCPOSBuilder {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	widthBytes := (width + 7) / 8
	xL := byte(widthBytes % 256)
	xH := byte(widthBytes / 256)
	yL := byte(height % 256)
	yH := byte(height / 256)

	// GS v 0 0 xL xH yL yH
	header := []byte{0x1D, 0x76, 0x30, 0x00, xL, xH, yL, yH}
	b.buf.Write(header)

	for y := 0; y < height; y++ {
		for xByte := 0; xByte < widthBytes; xByte++ {
			var bByte byte = 0
			for bit := 0; bit < 8; bit++ {
				x := xByte*8 + bit
				if x < width {
					c := color.GrayModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.Gray)
					// If pixel is dark (< 128 luminance), set bit to 1 (print black dot)
					if c.Y < 128 {
						bByte |= (1 << (7 - bit))
					}
				}
			}
			b.buf.WriteByte(bByte)
		}
	}

	return b
}

func (b *ESCPOSBuilder) Bytes() []byte {
	return b.buf.Bytes()
}
