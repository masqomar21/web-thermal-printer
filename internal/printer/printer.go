package printer

import (
	"fmt"
	"log"
	"net"
	"os"
)

// Printer represents a generic printer device interface
type Printer interface {
	Open() error
	Write(data []byte) (int, error)
	Close() error
	TestPrint() error
	Name() string
}


// NetPrinter communicates over TCP (e.g. 192.168.1.200:9100)
type NetPrinter struct {
	addr string
	conn net.Conn
}

func NewNetPrinter(addr string) *NetPrinter {
	return &NetPrinter{addr: addr}
}

func (p *NetPrinter) Name() string {
	return fmt.Sprintf("Network Printer (%s)", p.addr)
}

func (p *NetPrinter) Open() error {
	conn, err := net.Dial("tcp", p.addr)
	if err != nil {
		return fmt.Errorf("failed to connect to network printer at %s: %w", p.addr, err)
	}
	p.conn = conn
	return nil
}

func (p *NetPrinter) Write(data []byte) (int, error) {
	if p.conn == nil {
		if err := p.Open(); err != nil {
			return 0, err
		}
	}
	return p.conn.Write(data)
}

func (p *NetPrinter) Close() error {
	if p.conn != nil {
		err := p.conn.Close()
		p.conn = nil
		return err
	}
	return nil
}

func (p *NetPrinter) TestPrint() error {
	if err := p.Open(); err != nil {
		return err
	}
	defer p.Close()

	b := NewESCPOSBuilder()
	b.AlignCenter().SetFontSize(1, 1).SetBold(true).TextLn("PRINTER TEST READY!").NewLine(3).CutPaper()
	_, err := p.Write(b.Bytes())
	return err
}

// FilePrinter communicates directly with raw device nodes (e.g. /dev/usb/lp0, /dev/ttyUSB0, COM1)
type FilePrinter struct {
	path string
	file *os.File
}

func NewFilePrinter(path string) *FilePrinter {
	if path == "" {
		path = "/dev/usb/lp0"
	}
	return &FilePrinter{path: path}
}

func (p *FilePrinter) Name() string {
	return fmt.Sprintf("Raw File/Device Printer (%s)", p.path)
}

func (p *FilePrinter) Open() error {
	f, err := os.OpenFile(p.path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open printer device %s: %w", p.path, err)
	}
	p.file = f
	return nil
}

func (p *FilePrinter) Write(data []byte) (int, error) {
	if p.file == nil {
		if err := p.Open(); err != nil {
			return 0, err
		}
	}
	return p.file.Write(data)
}

func (p *FilePrinter) Close() error {
	if p.file != nil {
		err := p.file.Close()
		p.file = nil
		return err
	}
	return nil
}

func (p *FilePrinter) TestPrint() error {
	if err := p.Open(); err != nil {
		return err
	}
	defer p.Close()

	b := NewESCPOSBuilder()
	b.AlignCenter().SetFontSize(1, 1).SetBold(true).TextLn("PRINTER TEST READY!").NewLine(3).CutPaper()
	_, err := p.Write(b.Bytes())
	return err
}

// DummyPrinter for fallback testing or dry-run when printer device is unavailable
type DummyPrinter struct{}

func NewDummyPrinter() *DummyPrinter {
	return &DummyPrinter{}
}

func (p *DummyPrinter) Name() string { return "Dummy Virtual Printer (Console log)" }
func (p *DummyPrinter) Open() error  { return nil }
func (p *DummyPrinter) Close() error { return nil }
func (p *DummyPrinter) Write(data []byte) (int, error) {
	log.Printf("🖨️ [DummyPrinter] Bytes written: %d", len(data))
	return len(data), nil
}
func (p *DummyPrinter) TestPrint() error {
	log.Println("🖨️ [DummyPrinter] Test print succeeded")
	return nil
}
