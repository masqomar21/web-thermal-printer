//go:build js && wasm

package printer

import (
	"errors"
	"fmt"
	"syscall/js"
)

// WASMPrinter handles printing via Web Serial or Web USB JS API
type WASMPrinter struct {
	mode  string   // "serial" or "usb"
	port  js.Value // SerialPort or USBDevice object
	epNum int      // Endpoint number (for Web USB)
}

func NewWASMPrinter() *WASMPrinter {
	return &WASMPrinter{
		mode: "none",
	}
}

func (p *WASMPrinter) Name() string {
	return fmt.Sprintf("WASM Web Printer (%s)", p.mode)
}

func (p *WASMPrinter) Mode() string {
	return p.mode
}

func (p *WASMPrinter) SetSerialPort(port js.Value) {
	p.mode = "serial"
	p.port = port
}

func (p *WASMPrinter) SetUSBDevice(device js.Value, endpointNumber int) {
	p.mode = "usb"
	p.port = device
	p.epNum = endpointNumber
}

func (p *WASMPrinter) Open() error {
	if p.port.IsUndefined() || p.port.IsNull() {
		return errors.New("no Web Serial port or Web USB device selected")
	}
	return nil
}

func (p *WASMPrinter) Write(data []byte) (int, error) {
	if p.mode == "none" || p.port.IsUndefined() || p.port.IsNull() {
		return 0, errors.New("printer connection not initialized")
	}

	// Copy Go byte slice to JS Uint8Array
	uint8Array := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(uint8Array, data)

	if p.mode == "serial" {
		return p.writeSerial(uint8Array, len(data))
	} else if p.mode == "usb" {
		return p.writeUSB(uint8Array, len(data))
	}

	return 0, fmt.Errorf("unsupported WASM printer mode: %s", p.mode)
}

func (p *WASMPrinter) writeSerial(uint8Array js.Value, length int) (int, error) {
	writable := p.port.Get("writable")
	if writable.IsUndefined() || writable.IsNull() {
		return 0, errors.New("serial port is not open or not writable")
	}

	writer := writable.Call("getWriter")

	promise := writer.Call("write", uint8Array)

	done := make(chan error, 1)

	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		done <- nil
		return nil
	})
	defer thenFunc.Release()

	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		errMsg := "unknown serial write error"
		if len(args) > 0 && !args[0].IsUndefined() {
			errMsg = args[0].String()
		}
		done <- errors.New(errMsg)
		return nil
	})
	defer catchFunc.Release()

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	err := <-done
	writer.Call("releaseLock")

	if err != nil {
		return 0, fmt.Errorf("serial write failed: %w", err)
	}
	return length, nil
}

func (p *WASMPrinter) writeUSB(uint8Array js.Value, length int) (int, error) {
	promise := p.port.Call("transferOut", p.epNum, uint8Array)

	done := make(chan error, 1)

	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		done <- nil
		return nil
	})
	defer thenFunc.Release()

	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		errMsg := "unknown USB transferOut error"
		if len(args) > 0 && !args[0].IsUndefined() {
			errMsg = args[0].String()
		}
		done <- errors.New(errMsg)
		return nil
	})
	defer catchFunc.Release()

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	err := <-done
	if err != nil {
		return 0, fmt.Errorf("USB transferOut failed: %w", err)
	}
	return length, nil
}

func (p *WASMPrinter) Close() error {
	if p.mode == "serial" && !p.port.IsUndefined() && !p.port.IsNull() {
		closePromise := p.port.Call("close")
		done := make(chan struct{}, 1)
		fn := js.FuncOf(func(this js.Value, args []js.Value) any {
			done <- struct{}{}
			return nil
		})
		defer fn.Release()
		closePromise.Call("then", fn).Call("catch", fn)
		<-done
	}
	p.mode = "none"
	p.port = js.Null()
	return nil
}

func (p *WASMPrinter) TestPrint() error {
	b := NewESCPOSBuilder()
	b.AlignCenter().SetFontSize(1, 1).SetBold(true).TextLn("WASM PRINTER TEST READY!").NewLine(3).CutPaper()
	_, err := p.Write(b.Bytes())
	return err
}
