package renderer

import (
	"github.com/masqomar21/antrean-ticket-printer/internal/model"
)

// RenderTextTicket formats ticket data into ESC/POS text command bytes
func RenderTextTicket(data model.TicketData) []byte {
	return RenderTextDocument(data.ToDocument())
}

