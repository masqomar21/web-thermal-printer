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
