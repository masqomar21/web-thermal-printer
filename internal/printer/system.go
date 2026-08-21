package printer

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
)

// SystemPrinter sends raw ESC/POS bytes through OS print spooler (lpr on Unix, raw print on Windows)
type SystemPrinter struct {
	printerName string
	buf         bytes.Buffer
}

func NewSystemPrinter(printerName string) *SystemPrinter {
	return &SystemPrinter{printerName: printerName}
}

func (p *SystemPrinter) Name() string {
	if p.printerName != "" {
		return fmt.Sprintf("System Spooler Printer (%s)", p.printerName)
	}
	return "System Default Printer"
}

func (p *SystemPrinter) Open() error {
	p.buf.Reset()
	return nil
}

func (p *SystemPrinter) Write(data []byte) (int, error) {
	return p.buf.Write(data)
}

func (p *SystemPrinter) Close() error {
	if p.buf.Len() == 0 {
		return nil
	}
	defer p.buf.Reset()

	switch runtime.GOOS {
	case "windows":
		// Print raw file to Windows spooler using PowerShell or Out-Printer
		cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Get-Content -Raw | Out-Printer -Name '%s'", p.printerName))
		cmd.Stdin = &p.buf
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("windows spooler print failed: %w", err)
		}
	default:
		// Linux / macOS: lpr command
		args := []string{"-o", "raw"}
		if p.printerName != "" {
			args = append(args, "-P", p.printerName)
		}
		cmd := exec.Command("lpr", args...)
		cmd.Stdin = &p.buf
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("lpr spooler print failed: %w", err)
		}
	}

	return nil
}

func (p *SystemPrinter) TestPrint() error {
	if err := p.Open(); err != nil {
		return err
	}

	b := NewESCPOSBuilder()
	b.AlignCenter().SetFontSize(1, 1).SetBold(true).TextLn("PRINTER TEST READY!").NewLine(3).CutPaper()
	_, _ = p.Write(b.Bytes())

	return p.Close()
}
