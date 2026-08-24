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

func (b *ESCPOSBuilder) Divider(char string, lineLength int) *ESCPOSBuilder {
	if char == "" {
		char = "-"
	}
	if lineLength <= 0 {
		lineLength = 32 // default 58mm text width
	}
	line := ""
	for len(line) < lineLength {
		line += char
	}
	if len(line) > lineLength {
		line = line[:lineLength]
	}
	b.buf.WriteString(line + "\n")
	return b
}

func (b *ESCPOSBuilder) TableLine(left, right string, totalWidth int) *ESCPOSBuilder {
	if totalWidth <= 0 {
		totalWidth = 32
	}
	spaceNeeded := totalWidth - len(left) - len(right)
	if spaceNeeded <= 0 {
		b.buf.WriteString(left + " " + right + "\n")
		return b
	}
	spaces := ""
	for i := 0; i < spaceNeeded; i++ {
		spaces += " "
	}
	b.buf.WriteString(left + spaces + right + "\n")
	return b
}

func (b *ESCPOSBuilder) QRCode(content string, moduleSize int) *ESCPOSBuilder {
	if content == "" {
		return b
	}
	if moduleSize <= 0 {
		moduleSize = 4
	}
	if moduleSize > 16 {
		moduleSize = 16
	}

	dataLen := len(content) + 3
	pL := byte(dataLen % 256)
	pH := byte(dataLen / 256)

	// 1. Model: GS ( k 4 0 49 65 50 0 (Model 2)
	b.buf.Write([]byte{0x1D, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00})
	// 2. Size: GS ( k 3 0 49 67 <size>
	b.buf.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, byte(moduleSize)})
	// 3. Error Correction L: GS ( k 3 0 49 69 48
	b.buf.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x45, 0x30})
	// 4. Store Data: GS ( k pL pH 49 80 48 <data>
	b.buf.Write([]byte{0x1D, 0x28, 0x6B, pL, pH, 0x31, 0x50, 0x30})
	b.buf.WriteString(content)
	// 5. Print QR: GS ( k 3 0 49 81 48
	b.buf.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30})
	b.buf.WriteByte('\n')

	return b
}

func (b *ESCPOSBuilder) BarcodeCODE128(content string) *ESCPOSBuilder {
	if content == "" {
		return b
	}
	// GS h 60 (Height)
	b.buf.Write([]byte{0x1D, 0x68, 0x3C})
	// GS w 2 (Width)
	b.buf.Write([]byte{0x1D, 0x77, 0x02})
	// GS H 2 (Text below)
	b.buf.Write([]byte{0x1D, 0x48, 0x02})

	// CODE128 command: GS k 73 <length> {CODE B} <content>
	encoded := "{B" + content
	dataLen := byte(len(encoded))
	b.buf.Write([]byte{0x1D, 0x6B, 0x49, dataLen})
	b.buf.WriteString(encoded)
	b.buf.WriteByte('\n')

	return b
}

func (b *ESCPOSBuilder) Bytes() []byte {
	return b.buf.Bytes()
}

