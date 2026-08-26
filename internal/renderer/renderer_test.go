package renderer

import (
	"testing"

	"github.com/masqomar21/antrean-ticket-printer/internal/model"
)

func TestRenderTextTicket(t *testing.T) {
	ticket := model.TicketData{
		Instansi:     "Puskesmas Maju Jaya",
		Loket:        "Loket 1",
		NomorAntrean: "A-001",
		Tanggal:      "2026-08-21",
		Jam:          "14:00",
	}

	bytes := RenderTextTicket(ticket)
	if len(bytes) == 0 {
		t.Fatal("expected non-empty ESC/POS text bytes")
	}
}

func TestRenderImageTicket(t *testing.T) {
	ticket := model.TicketData{
		Instansi:     "Puskesmas Maju Jaya",
		Loket:        "Loket 1",
		NomorAntrean: "A-001",
		Tanggal:      "2026-08-21",
		Jam:          "14:00",
	}

	bytes := RenderImageTicket(ticket, 576)
	if len(bytes) == 0 {
		t.Fatal("expected non-empty ESC/POS raster image bytes")
	}
}

func TestRenderTextDocument(t *testing.T) {
	doc := model.PrintDocument{
		Header: &model.HeaderConfig{
			Title:    "TOKO MAJU BERSAMA",
			Instansi: "JL. MERDEKA NO. 123",
			Align:    "center",
		},
		Fields: []model.KeyValueField{
			{Key: "No. Struk", Value: "TRX-998822"},
			{Key: "Kasir", Value: "Budi"},
		},
		Items: []model.TableItem{
			{Name: "Kopi Hitam", Qty: "2", Price: "15.000", Total: "30.000"},
			{Name: "Roti Bakar", Qty: "1", Price: "20.000", Total: "20.000"},
		},
		Totals: []model.KeyValueField{
			{Key: "Total", Value: "50.000", BoldKey: true, BoldValue: true},
		},
		Barcode: &model.BarcodeConfig{
			Type:    "QR",
			Content: "TRX-998822",
			Align:   "center",
		},
		Footer: &model.FooterConfig{
			Lines:   []string{"Terima Kasih atas Kunjungan Anda"},
			Align:   "center",
			Divider: true,
		},
	}

	bytes := RenderTextDocument(doc)
	if len(bytes) == 0 {
		t.Fatal("expected non-empty ESC/POS generic document bytes")
	}
}

