package renderer

import (
	"github.com/masqomar21/antrean-ticket-printer/internal/model"
	"github.com/masqomar21/antrean-ticket-printer/internal/printer"
)

// RenderTextTicket formats ticket data into ESC/POS text command bytes
func RenderTextTicket(data model.TicketData) []byte {
	b := printer.NewESCPOSBuilder()

	// Title / Instansi (if present)
	if data.Instansi != "" {
		b.AlignCenter().SetFontSize(0, 0).SetBold(true).TextLn(data.Instansi)
	}

	// Loket
	b.AlignCenter().SetFontSize(0, 1).SetBold(false).TextLn("Loket : " + data.Loket)
	b.NewLine(1)

	// Nomor Antrean (Large font: 2x width, 2x height)
	b.AlignCenter().SetFontSize(2, 2).SetBold(true).TextLn(data.NomorAntrean)
	b.NewLine(1)

	// Tanggal & Jam
	b.AlignCenter().SetFontSize(0, 0).SetBold(false).
		TextLn("Tanggal : " + data.Tanggal).
		TextLn("Jam : " + data.Jam)

	if data.Layanan != "" {
		b.TextLn("Layanan : " + data.Layanan)
	}

	// Feed & Cut Paper
	b.NewLine(3).CutPaper()

	return b.Bytes()
}
