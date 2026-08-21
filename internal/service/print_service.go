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
	start := time.Now()
	log.Printf("📄 Processing ticket for Loket: %s, Nomor: %s", data.Loket, data.NomorAntrean)

	var payload []byte
	mode := strings.ToLower(s.cfg.RenderMode)

	switch mode {
	case "image", "raster":
		log.Println("🖼️ Rendering ticket in Image/Raster mode...")
		payload = renderer.RenderImageTicket(data, s.cfg.PaperWidthDots)
	case "both":
		log.Println("🖨️ Rendering ticket in Text & Graphic dual mode...")
		payload = renderer.RenderTextTicket(data)
	default: // "text"
		log.Println("📝 Rendering ticket in Direct ESC/POS Text mode...")
		payload = renderer.RenderTextTicket(data)
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
