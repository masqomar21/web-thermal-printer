package renderer

import (
	"bytes"

	"github.com/masqomar21/antrean-ticket-printer/internal/model"
	"github.com/masqomar21/antrean-ticket-printer/internal/printer"
)

// RenderTextDocument formats a generic PrintDocument into ESC/POS command bytes
func RenderTextDocument(doc model.PrintDocument) []byte {
	b := printer.NewESCPOSBuilder()
	lineWidth := 32 // Standard 58mm text width

	// 1. Header Section
	if doc.Header != nil {
		h := doc.Header
		setAlign(b, h.Align)

		if h.Title != "" {
			b.SetFontSize(1, 1).SetBold(true).TextLn(h.Title)
		}
		if h.Instansi != "" {
			b.SetFontSize(0, 0).SetBold(true).TextLn(h.Instansi)
		}
		if h.Subtitle != "" {
			b.SetFontSize(0, 0).SetBold(false).TextLn(h.Subtitle)
		}
		if h.Address != "" {
			b.SetFontSize(0, 0).SetBold(false).TextLn(h.Address)
		}
		if h.Phone != "" {
			b.SetFontSize(0, 0).SetBold(false).TextLn("Telp: " + h.Phone)
		}
		b.NewLine(1)
	}

	// 2. Key-Value Fields Section
	if len(doc.Fields) > 0 {
		for _, f := range doc.Fields {
			setAlign(b, f.Align)
			b.SetBold(f.BoldKey || f.BoldValue)

			if f.Align == "center" || f.Align == "right" {
				if f.Key != "" {
					b.TextLn(f.Key + " : " + f.Value)
				} else {
					b.TextLn(f.Value)
				}
			} else {
				if f.Key != "" {
					b.TableLine(f.Key, f.Value, lineWidth)
				} else {
					b.TextLn(f.Value)
				}
			}
		}
		b.NewLine(1)
	}

	// 3. Table / Items Section (Struk / Invoice)
	if len(doc.Items) > 0 {
		b.Divider("-", lineWidth)
		for _, item := range doc.Items {
			b.AlignLeft().SetBold(true).TextLn(item.Name)
			if item.Qty != "" || item.Price != "" || item.Total != "" {
				subLeft := "  " + item.Qty + " x " + item.Price
				b.TableLine(subLeft, item.Total, lineWidth)
			}
			if item.Details != "" {
				b.AlignLeft().SetBold(false).TextLn("  " + item.Details)
			}
		}
		b.Divider("-", lineWidth)
	}

	// 4. Totals Section
	if len(doc.Totals) > 0 {
		for _, t := range doc.Totals {
			b.SetBold(t.BoldKey || t.BoldValue)
			b.TableLine(t.Key, t.Value, lineWidth)
		}
		b.Divider("=", lineWidth)
	}

	// 5. Barcode / QR Code Section
	if doc.Barcode != nil {
		setAlign(b, doc.Barcode.Align)
		b.NewLine(1)
		switch doc.Barcode.Type {
		case "QR", "qr":
			b.QRCode(doc.Barcode.Content, doc.Barcode.Size)
		case "CODE128", "code128":
			b.BarcodeCODE128(doc.Barcode.Content)
		default:
			b.QRCode(doc.Barcode.Content, doc.Barcode.Size)
		}
	}

	// 6. Footer Section
	if doc.Footer != nil {
		f := doc.Footer
		setAlign(b, f.Align)
		b.NewLine(1)

		if f.Divider {
			b.Divider("-", lineWidth)
		}
		for _, line := range f.Lines {
			b.TextLn(line)
		}
	}

	// 7. Feed & Cut Options
	feed := 3
	cut := true
	if doc.Options != nil {
		if doc.Options.FeedLines > 0 {
			feed = doc.Options.FeedLines
		}
		cut = doc.Options.CutPaper
	}

	b.NewLine(feed)
	if cut {
		b.CutPaper()
	}

	return b.Bytes()
}

func setAlign(b *printer.ESCPOSBuilder, align string) {
	switch align {
	case "center":
		b.AlignCenter()
	case "right":
		b.AlignRight()
	default:
		b.AlignLeft()
	}
}

// RenderImageDocument converts document to raster image ESC/POS bytes (placeholder / image adapter)
func RenderImageDocument(doc model.PrintDocument, widthDots int) []byte {
	// Fallback to text rendering or raster generator
	var buf bytes.Buffer
	textBytes := RenderTextDocument(doc)
	buf.Write(textBytes)
	return buf.Bytes()
}
