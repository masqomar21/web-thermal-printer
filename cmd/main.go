package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/masqomar21/antrean-ticket-printer/internal/config"
	"github.com/masqomar21/antrean-ticket-printer/internal/model"
	"github.com/masqomar21/antrean-ticket-printer/internal/service"
	"github.com/masqomar21/antrean-ticket-printer/internal/socket"
)

func main() {
	printBanner()

	// 1. Load Application Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	log.Printf("⚙️ Configuration Loaded: Socket URL = %s | Topic = %s", cfg.Socket.URL, cfg.Socket.TopicPrintNomorAntrean)
	log.Printf("🖨️ Printer Config: Type = %s | Render Mode = %s", cfg.Printer.Type, cfg.Printer.RenderMode)

	// 2. Initialize Print Service
	printSvc, err := service.NewPrintService(cfg.Printer)
	if err != nil {
		log.Fatalf("❌ Failed to initialize printer service: %v", err)
	}

	// 3. Ensure Printer is Ready
	go printSvc.EnsurePrinterReady()

	// 4. Initialize Socket.IO Client
	client := socket.NewClient(cfg.Socket, func(data model.TicketData) error {
		return printSvc.PrintTicket(data)
	})

	// 5. Start Socket Client Connection
	client.Start()

	// 6. Wait for Shutdown Signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	log.Println("🚀 Antrean Ticket Printer Service is running. Press Ctrl+C to stop.")
	sig := <-sigChan
	log.Printf("🛑 Received signal %v. Shutting down gracefully...", sig)

	client.Stop()
	log.Println("👋 Goodbye!")
}

func printBanner() {
	banner := `
=================================================
  🖨️  ANTREAN TICKET PRINTER - GO EDITION v1.0  
  ⚡ Fast, Lightweight & Native Thermal Printing  
=================================================
`
	fmt.Println(banner)
}
