package model

// PrintDocument represents a general-purpose printable document (receipt, invoice, queue ticket, custom report, etc.)
type PrintDocument struct {
	Header  *HeaderConfig   `json:"header,omitempty"`
	Fields  []KeyValueField `json:"fields,omitempty"`
	Items   []TableItem     `json:"items,omitempty"`
	Totals  []KeyValueField `json:"totals,omitempty"`
	Barcode *BarcodeConfig  `json:"barcode,omitempty"`
	Footer  *FooterConfig   `json:"footer,omitempty"`
	Options *PrintOptions   `json:"options,omitempty"`
}

type HeaderConfig struct {
	Title     string `json:"title,omitempty"`
	Subtitle  string `json:"subtitle,omitempty"`
	Instansi  string `json:"instansi,omitempty"`
	Address   string `json:"address,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Align     string `json:"align,omitempty"`     // left, center, right
	LargeFont bool   `json:"large_font,omitempty"`
}

type KeyValueField struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	BoldKey   bool   `json:"bold_key,omitempty"`
	BoldValue bool   `json:"bold_value,omitempty"`
	Align     string `json:"align,omitempty"` // left, center, right
}

type TableItem struct {
	Name    string `json:"name"`
	Qty     string `json:"qty,omitempty"`
	Price   string `json:"price,omitempty"`
	Total   string `json:"total,omitempty"`
	Details string `json:"details,omitempty"`
}

type BarcodeConfig struct {
	Type    string `json:"type"` // QR, CODE128, EAN13
	Content string `json:"content"`
	Size    int    `json:"size,omitempty"`
	Align   string `json:"align,omitempty"` // left, center, right
}

type FooterConfig struct {
	Lines   []string `json:"lines,omitempty"`
	Align   string   `json:"align,omitempty"` // left, center, right
	Divider bool     `json:"divider,omitempty"`
}

type PrintOptions struct {
	CutPaper       bool   `json:"cut_paper,omitempty"`
	FeedLines      int    `json:"feed_lines,omitempty"`
	RenderMode     string `json:"render_mode,omitempty"` // text, image, raster
	PaperWidthDots int    `json:"paper_width_dots,omitempty"`
}

// TicketDataToDocument converts legacy antrean TicketData to generic PrintDocument
func TicketDataToDocument(t TicketData) PrintDocument {
	doc := PrintDocument{
		Header: &HeaderConfig{
			Instansi: t.Instansi,
			Align:    "center",
		},
		Fields: []KeyValueField{},
		Options: &PrintOptions{
			CutPaper:  true,
			FeedLines: 3,
		},
	}

	if t.Loket != "" {
		doc.Fields = append(doc.Fields, KeyValueField{
			Key:   "Loket",
			Value: t.Loket,
			Align: "center",
		})
	}

	if t.NomorAntrean != "" {
		doc.Fields = append(doc.Fields, KeyValueField{
			Key:       "Nomor Antrean",
			Value:     t.NomorAntrean,
			BoldValue: true,
			Align:     "center",
		})
	}

	if t.Tanggal != "" {
		doc.Fields = append(doc.Fields, KeyValueField{
			Key:   "Tanggal",
			Value: t.Tanggal,
			Align: "center",
		})
	}

	if t.Jam != "" {
		doc.Fields = append(doc.Fields, KeyValueField{
			Key:   "Jam",
			Value: t.Jam,
			Align: "center",
		})
	}

	if t.Layanan != "" {
		doc.Fields = append(doc.Fields, KeyValueField{
			Key:   "Layanan",
			Value: t.Layanan,
			Align: "center",
		})
	}

	return doc
}
