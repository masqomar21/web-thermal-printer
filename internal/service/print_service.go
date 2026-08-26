package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/masqomar21/antrean-ticket-printer/internal/config"
	"github.com/masqomar21/antrean-ticket-printer/internal/model"
	"github.com/masqomar21/antrean-ticket-printer/internal/printer"
	"github.com/masqomar21/antrean-ticket-printer/internal/renderer"
)

type PrintService struct {
	cfg     config.PrinterConfig
	printer printer.Printer
}

func NewPrintService(cfg config.PrinterConfig) (*PrintService, error) {
	p, err := printer.NewPrinter(cfg)
	if err != nil {
		return nil, err
	}
	return &PrintService{
		cfg:     cfg,
		printer: p,
	}, nil
}

// EnsurePrinterReady tests the printer until it responds, similar to testPrint() loop in Node.js index.js
func (s *PrintService) EnsurePrinterReady() {
	log.Printf("🔍 Checking printer connection [%s]...", s.printer.Name())
	for {
		err := s.printer.TestPrint()
		if err == nil {
			log.Println("🖨️ Printer ready!")
			return
		}

		log.Printf("⚠️ Printer test failed (%v). Retrying in 2 seconds...", err)
		time.Sleep(2 * time.Second)

		// Try re-creating printer instance in case of USB disconnect
		if p, err := printer.NewPrinter(s.cfg); err == nil {
			s.printer = p
		}
	}
}

func (s *PrintService) PrintTicket(data model.TicketData) error {
	return s.PrintDocument(data.ToDocument())
}

func (s *PrintService) PrintDocument(doc model.PrintDocument) error {
	start := time.Now()
	log.Println("📄 Processing generic print document...")

	var payload []byte
	mode := strings.ToLower(s.cfg.RenderMode)

	if doc.Options != nil && doc.Options.RenderMode != "" {
		mode = strings.ToLower(doc.Options.RenderMode)
	}

	switch mode {
	case "image", "raster":
		widthDots := s.cfg.PaperWidthDots
		if doc.Options != nil && doc.Options.PaperWidthDots > 0 {
			widthDots = doc.Options.PaperWidthDots
		}
		log.Println("🖼️ Rendering document in Image/Raster mode...")
		payload = renderer.RenderImageDocument(doc, widthDots)
	default: // "text"
		log.Println("📝 Rendering document in ESC/POS Text mode...")
		payload = renderer.RenderTextDocument(doc)
	}

	if err := s.printer.Open(); err != nil {
		return fmt.Errorf("failed to open printer: %w", err)
	}
	defer s.printer.Close()

	if _, err := s.printer.Write(payload); err != nil {
		return fmt.Errorf("failed to write print data: %w", err)
	}

	elapsed := time.Since(start)
	log.Printf("✅ Print finished in %v", elapsed)
	return nil
}

