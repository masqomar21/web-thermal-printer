package printer

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// USBPrinter represents a thermal printer connected over USB or raw USB endpoint
type USBPrinter struct {
	vendorID     uint16
	productID    uint16
	path         string
	resolvedPath string
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
	if p.resolvedPath != "" {
		return fmt.Sprintf("USB Thermal Printer (%s)", p.resolvedPath)
	}
	return "Auto-Detect USB Thermal Printer"
}

func (p *USBPrinter) Open() error {
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
		// 1. Auto-discover Windows USB Raw Device Interface Paths (\??\USB#VID_... / \\?\USB#...)
		usbPaths := p.findWindowsUSBPaths()
		candidatePaths = append(candidatePaths, usbPaths...)

		// 2. Standard Windows RAW USB printer port paths
		candidatePaths = append(candidatePaths,
			"\\\\.\\USB001", "\\\\.\\USB002", "\\\\.\\USB003", "\\\\.\\USB004", "\\\\.\\USB005",
			"\\\\.\\LPT1", "\\\\.\\LPT2",
		)

		// 3. Serial / COM ports as last resort fallback
		if p.path == "" {
			candidatePaths = append(candidatePaths,
				"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
			)
		}

	case "darwin": // macOS
		candidatePaths = append(candidatePaths,
			"/dev/cu.usbserial", "/dev/tty.usbserial",
		)
	}

	var lastErr error
	for _, path := range candidatePaths {
		// Test writing 0 bytes or opening to verify connection
		testBytes := []byte{}
		_, err := writeUSBDataWindows(path, testBytes)
		if err == nil {
			log.Printf("✅ Thermal printer connected successfully at %s", path)
			p.resolvedPath = path
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("unable to open USB/Raw printer (tried paths: %v): %w", candidatePaths, lastErr)
}

func (p *USBPrinter) findWindowsUSBPaths() []string {
	paths := []string{}
	var targetVidPid string
	if p.vendorID != 0 && p.productID != 0 {
		targetVidPid = fmt.Sprintf("VID_%04X&PID_%04X", p.vendorID, p.productID)
	}

	psScript := `Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Enum\USB\*\*\Device Parameters' -ErrorAction SilentlyContinue | Where-Object { $_.SymbolicName } | Select-Object -ExpandProperty SymbolicName`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			winPath := strings.Replace(line, `\??\`, `\\?\`, 1)

			if targetVidPid != "" {
				if strings.Contains(strings.ToUpper(winPath), strings.ToUpper(targetVidPid)) {
					paths = append(paths, winPath)
				}
			} else {
				upperPath := strings.ToUpper(winPath)
				if strings.Contains(upperPath, "VID_0FE6") || strings.Contains(upperPath, "VID_04B8") ||
					strings.Contains(upperPath, "VID_1504") || strings.Contains(upperPath, "VID_0416") ||
					strings.Contains(upperPath, "VID_0DD4") || strings.Contains(upperPath, "USBPRINT") {
					paths = append([]string{winPath}, paths...)
				} else {
					paths = append(paths, winPath)
				}
			}
		}
	}

	return paths
}

func (p *USBPrinter) Write(data []byte) (int, error) {
	if p.resolvedPath == "" {
		if err := p.Open(); err != nil {
			return 0, err
		}
	}
	return writeUSBDataWindows(p.resolvedPath, data)
}

func (p *USBPrinter) Close() error {
	p.resolvedPath = ""
	return nil
}

func (p *USBPrinter) TestPrint() error {
	if err := p.Open(); err != nil {
		return err
	}

	b := NewESCPOSBuilder()
	b.AlignCenter().SetFontSize(1, 1).SetBold(true).TextLn("PRINTER TEST READY!").NewLine(3).CutPaper()
	_, err := p.Write(b.Bytes())
	return err
}
