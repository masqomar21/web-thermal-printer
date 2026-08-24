package model

// TicketData represents the payload received from socket event (e.g. antrean_print)
type TicketData struct {
	Instansi     string `json:"instansi,omitempty"`
	Layanan      string `json:"layanan,omitempty"`
	NomorAntrean string `json:"nomor_antrean"`
	Tanggal      string `json:"tanggal"`
	Jam          string `json:"jam"`
	Loket        string `json:"loket"`
}

// PrintStatusPayload represents status updates sent back to socket server
type PrintStatusPayload struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ToDocument converts TicketData into a generic PrintDocument
func (t TicketData) ToDocument() PrintDocument {
	return TicketDataToDocument(t)
}

