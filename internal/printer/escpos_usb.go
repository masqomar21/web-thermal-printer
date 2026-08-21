package printer

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

// USBPrinter represents a thermal printer connected over USB or raw USB endpoint
type USBPrinter struct {
	vendorID  uint16
	productID uint16
	path      string
	file      *os.File
}

func NewUSBPrinter(vendorID, productID uint16, rawPath string) *USBPrinter {
	return &USBPrinter{
		vendorID:  vendorID,
		productID: productID,
		path:      rawPath,
	}
}

func (p *USBPrinter) Name() string {
	if p.vendorID != 0 && p.productID != 0 {
		return fmt.Sprintf("USB Printer (VID:0x%04X PID:0x%04X)", p.vendorID, p.productID)
	}
	return "Auto-Detect USB Thermal Printer"
}

func (p *USBPrinter) Open() error {
	// Attempt candidate paths depending on OS
	candidatePaths := []string{}
	if p.path != "" {
		candidatePaths = append(candidatePaths, p.path)
	}

	switch runtime.GOOS {
	case "linux":
		candidatePaths = append(candidatePaths,
			"/dev/usb/lp0", "/dev/usb/lp1", "/dev/usb/lp2",
			"/dev/usblp0", "/dev/usblp1",
		)
	case "windows":
		candidatePaths = append(candidatePaths,
			"\\\\.\\LPT1", "\\\\.\\LPT2",
			"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		)
	case "darwin": // macOS
		candidatePaths = append(candidatePaths,
			"/dev/cu.usbserial", "/dev/tty.usbserial",
		)
	}

	var lastErr error
	for _, path := range candidatePaths {
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err == nil {
			log.Printf("✅ Thermal printer opened successfully at %s", path)
			p.file = f
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("unable to open USB/Raw printer (tried paths: %v): %w", candidatePaths, lastErr)
}

func (p *USBPrinter) Write(data []byte) (int, error) {
	if p.file == nil {
		if err := p.Open(); err != nil {
			return 0, err
		}
	}
	return p.file.Write(data)
}

func (p *USBPrinter) Close() error {
	if p.file != nil {
		err := p.file.Close()
		p.file = nil
		return err
	}
	return nil
}

func (p *USBPrinter) TestPrint() error {
	if err := p.Open(); err != nil {
		return err
	}
	defer p.Close()

	b := NewESCPOSBuilder()
	b.AlignCenter().SetFontSize(1, 1).SetBold(true).TextLn("PRINTER TEST READY!").NewLine(3).CutPaper()
	_, err := p.Write(b.Bytes())
	return err
}
