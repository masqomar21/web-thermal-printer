package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/masqomar21/antrean-ticket-printer/internal/config"
	"github.com/masqomar21/antrean-ticket-printer/internal/model"
	"github.com/masqomar21/antrean-ticket-printer/internal/service"
)

func main() {
	printHeader()

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to load configuration (.env / config.json): %v", err)
		log.Println("💡 Defaulting to fallback printer configuration (Type: usb, Render: text)")
		cfg = &config.AppConfig{
			Printer: config.PrinterConfig{
				Type:           "usb",
				RenderMode:     "text",
				PaperWidthDots: 576,
			},
		}
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		displayCurrentConfig(cfg.Printer)
		displayMenu()

		fmt.Print("👉 Pilihan Anda (0-5): ")
		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		fmt.Println()
		switch choice {
		case "1":
			runTestPrint(cfg.Printer, "text")
		case "2":
			runTestPrint(cfg.Printer, "image")
		case "3":
			runTestPrint(cfg.Printer, "both")
		case "4":
			overrideConfig(reader, &cfg.Printer)
		case "5":
			reloaded, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("❌ Gagal memuat ulang konfigurasi: %v\n", err)
			} else {
				cfg = reloaded
				fmt.Println("✅ Konfigurasi berhasil dimuat ulang dari file!")
			}
		case "0":
			fmt.Println("👋 Keluar dari Alat Pengujian Printer. Terima kasih!")
			return
		default:
			fmt.Println("⚠️ Pilihan tidak valid. Silakan masukkan angka 0 - 5.")
		}

		fmt.Println("\n" + strings.Repeat("-", 50))
	}
}

func printHeader() {
	fmt.Println(`
=================================================
  🧪 ANTREAN TICKET PRINTER - TEST UTILITY (GO)
  ⚡ Alat Pengujian Koneksi & Hasil Cetak Thermal
=================================================`)
}

func displayCurrentConfig(pCfg config.PrinterConfig) {
	fmt.Println("\n📌 KONFIGURASI PRINTER AKTIF SAAT INI:")
	fmt.Printf("   • Tipe Printer       : %s\n", strings.ToUpper(pCfg.Type))
	fmt.Printf("   • Mode Render Default : %s\n", strings.ToUpper(pCfg.RenderMode))
	if pCfg.Type == "net" || pCfg.Type == "network" || pCfg.Type == "tcp" {
		fmt.Printf("   • IP Address Printer  : %s\n", pCfg.IPAddress)
	}
	if pCfg.Type == "system" || pCfg.Type == "spooler" {
		fmt.Printf("   • Nama Printer OS     : %s\n", pCfg.SystemPrinter)
	}
	if pCfg.Type == "file" || pCfg.Type == "raw" || pCfg.Type == "device" || pCfg.Type == "usb" {
		if pCfg.SerialPort != "" {
			fmt.Printf("   • Port / Path Device  : %s\n", pCfg.SerialPort)
		} else {
			fmt.Println("   • Port / Path Device  : AUTO-DETECT (USB)")
		}
		if pCfg.VendorID != 0 || pCfg.ProductID != 0 {
			fmt.Printf("   • USB Vendor/Product  : 0x%04X : 0x%04X\n", pCfg.VendorID, pCfg.ProductID)
		}
	}
	fmt.Printf("   • Lebar Kertas Dots   : %d px\n", pCfg.PaperWidthDots)
}

func displayMenu() {
	fmt.Println("\n📋 MENU PENGUJIAN PRINTER:")
	fmt.Println("  [1] 📄 Cetak Sampel Tiket - Mode ESC/POS Text Direct (Cepat)")
	fmt.Println("  [2] 🖼️ Cetak Sampel Tiket - Mode Graphic / Raster Image (Native Go Image)")
	fmt.Println("  [3] 🖨️ Cetak Sampel Tiket - Mode Both / Dual (Text & Grafik)")
	fmt.Println("  [4] ⚙️ Ubah / Override Konfigurasi Printer (Sementara)")
	fmt.Println("  [5] 🔄 Reload Konfigurasi dari File (.env / config.json)")
	fmt.Println("  [0] 🚪 Keluar")
}

func runTestPrint(pCfg config.PrinterConfig, renderMode string) {
	activeCfg := pCfg
	activeCfg.RenderMode = renderMode

	fmt.Printf("🚀 Memulai Test Print [%s] dalam mode [%s]...\n", activeCfg.Type, strings.ToUpper(renderMode))

	svc, err := service.NewPrintService(activeCfg)
	if err != nil {
		fmt.Printf("❌ Gagal inisialisasi printer: %v\n", err)
		return
	}

	now := time.Now()
	sampleTicket := model.TicketData{
		Instansi:     "PUSKESMAS MAJU SEJAHTERA",
		Layanan:      "POLIKLINIK UMUM",
		NomorAntrean: "A-015",
		Tanggal:      now.Format("2006-01-02"),
		Jam:          now.Format("15:04:05"),
		Loket:        "LOKET 1",
	}

	start := time.Now()
	err = svc.PrintTicket(sampleTicket)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ TEST PRINT GAGAL (%v): %v\n", elapsed, err)
	} else {
		fmt.Printf("✅ TEST PRINT BERHASIL DITERIMA PRINTER! Selesai dalam %v\n", elapsed)
	}
}

func overrideConfig(reader *bufio.Reader, pCfg *config.PrinterConfig) {
	fmt.Println("\n⚙️ UBAH KONFIGURASI PRINTER SEMENTARA:")
	fmt.Println("   1. USB (Auto-detect / Vendor-Product ID)")
	fmt.Println("   2. Network (TCP / IP Printer)")
	fmt.Println("   3. System Spooler (Windows/OS Printer)")
	fmt.Println("   4. Raw File Device / Serial Port (e.g. COM3, /dev/usb/lp0)")
	fmt.Print("👉 Pilih Tipe Printer (1-4): ")

	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	switch choice {
	case "1":
		pCfg.Type = "usb"
		fmt.Print("   Masukkan Vendor ID Hex (misal: 0x0fe6 atau tekan Enter untuk default/0): ")
		vidStr, _ := reader.ReadString('\n')
		vidStr = strings.TrimSpace(vidStr)
		if vidStr != "" {
			vidStr = strings.TrimPrefix(vidStr, "0x")
			vidStr = strings.TrimPrefix(vidStr, "0X")
			if vid, err := strconv.ParseUint(vidStr, 16, 16); err == nil {
				pCfg.VendorID = uint16(vid)
			}
		}
		fmt.Print("   Masukkan Product ID Hex (misal: 0x811e atau tekan Enter untuk default/0): ")
		pidStr, _ := reader.ReadString('\n')
		pidStr = strings.TrimSpace(pidStr)
		if pidStr != "" {
			pidStr = strings.TrimPrefix(pidStr, "0x")
			pidStr = strings.TrimPrefix(pidStr, "0X")
			if pid, err := strconv.ParseUint(pidStr, 16, 16); err == nil {
				pCfg.ProductID = uint16(pid)
			}
		}

	case "2":
		pCfg.Type = "net"
		fmt.Print("   Masukkan IP Address & Port (contoh: 192.168.1.200:9100): ")
		ipInput, _ := reader.ReadString('\n')
		ipInput = strings.TrimSpace(ipInput)
		if ipInput != "" {
			pCfg.IPAddress = ipInput
		}

	case "3":
		pCfg.Type = "system"
		fmt.Print("   Masukkan Nama Printer OS (contoh: POS-58 atau Thermal Printer): ")
		nameInput, _ := reader.ReadString('\n')
		nameInput = strings.TrimSpace(nameInput)
		if nameInput != "" {
			pCfg.SystemPrinter = nameInput
		}

	case "4":
		pCfg.Type = "file"
		fmt.Print("   Masukkan Device Path / COM Port (contoh: COM3 atau /dev/usb/lp0): ")
		devInput, _ := reader.ReadString('\n')
		devInput = strings.TrimSpace(devInput)
		if devInput != "" {
			pCfg.SerialPort = devInput
		}

	default:
		fmt.Println("⚠️ Tipe printer tidak diubah.")
		return
	}

	fmt.Println("✅ Konfigurasi sementara berhasil diperbarui!")
}
