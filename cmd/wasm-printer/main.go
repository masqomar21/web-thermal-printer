//go:build js && wasm

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/masqomar21/antrean-ticket-printer/internal/model"
	"github.com/masqomar21/antrean-ticket-printer/internal/printer"
	"github.com/masqomar21/antrean-ticket-printer/internal/renderer"
)

var globalPrinter = printer.NewWASMPrinter()

func main() {
	c := make(chan struct{}, 0)

	sdkObj := js.Global().Get("Object").New()
	sdkObj.Set("connectSerial", js.FuncOf(connectSerial))
	sdkObj.Set("connectUSB", js.FuncOf(connectUSB))
	sdkObj.Set("setSerialPort", js.FuncOf(setSerialPort))
	sdkObj.Set("setUSBDevice", js.FuncOf(setUSBDevice))
	sdkObj.Set("printTicket", js.FuncOf(printTicket))
	sdkObj.Set("printTicketImage", js.FuncOf(printTicketImage))
	sdkObj.Set("printImageBase64", js.FuncOf(printImageBase64))
	sdkObj.Set("printRawBytes", js.FuncOf(printRawBytes))
	sdkObj.Set("testPrint", js.FuncOf(testPrint))
	sdkObj.Set("close", js.FuncOf(closePrinter))
	sdkObj.Set("getStatus", js.FuncOf(getStatus))

	js.Global().Set("ThermalPrinterWASM", sdkObj)

	fmt.Println("🚀 Thermal Printer Go WASM Gateway Initialized!")
	<-c
}

// Helper to create JS Promise
func newPromise(executor func(resolve, reject js.Value)) js.Value {
	promiseCtor := js.Global().Get("Promise")
	return promiseCtor.New(js.FuncOf(func(this js.Value, args []js.Value) any {
		resolve := args[0]
		reject := args[1]
		go executor(resolve, reject)
		return nil
	}))
}

func connectSerial(this js.Value, args []js.Value) any {
	baudRate := 9600
	if len(args) > 0 && !args[0].IsUndefined() && args[0].Type() == js.TypeNumber {
		baudRate = args[0].Int()
	}

	return newPromise(func(resolve, reject js.Value) {
		nav := js.Global().Get("navigator")
		if nav.Get("serial").IsUndefined() {
			reject.Invoke("Web Serial API is not supported in this browser.")
			return
		}

		// Request serial port
		portPromise := nav.Get("serial").Call("requestPort")
		port, err := awaitPromise(portPromise)
		if err != nil {
			reject.Invoke(fmt.Sprintf("Failed to select serial port: %v", err))
			return
		}

		// Open port
		openOpts := js.Global().Get("Object").New()
		openOpts.Set("baudRate", baudRate)

		openPromise := port.Call("open", openOpts)
		_, err = awaitPromise(openPromise)
		if err != nil {
			reject.Invoke(fmt.Sprintf("Failed to open serial port: %v", err))
			return
		}

		globalPrinter.SetSerialPort(port)
		resolve.Invoke(js.ValueOf(true))
	})
}

func connectUSB(this js.Value, args []js.Value) any {
	endpoint := 1
	if len(args) > 0 && !args[0].IsUndefined() && args[0].Type() == js.TypeNumber {
		endpoint = args[0].Int()
	}

	return newPromise(func(resolve, reject js.Value) {
		nav := js.Global().Get("navigator")
		if nav.Get("usb").IsUndefined() {
			reject.Invoke("Web USB API is not supported in this browser.")
			return
		}

		reqOpts := js.Global().Get("Object").New()
		filters := js.Global().Get("Array").New()
		reqOpts.Set("filters", filters)

		devPromise := nav.Get("usb").Call("requestDevice", reqOpts)
		device, err := awaitPromise(devPromise)
		if err != nil {
			reject.Invoke(fmt.Sprintf("Failed to select USB device: %v", err))
			return
		}

		// Open & claim device
		_, err = awaitPromise(device.Call("open"))
		if err != nil {
			reject.Invoke(fmt.Sprintf("USB device open failed: %v", err))
			return
		}

		_, err = awaitPromise(device.Call("selectConfiguration", 1))
		if err != nil {
			reject.Invoke(fmt.Sprintf("USB selectConfiguration failed: %v", err))
			return
		}

		_, err = awaitPromise(device.Call("claimInterface", 0))
		if err != nil {
			reject.Invoke(fmt.Sprintf("USB claimInterface failed: %v", err))
			return
		}

		globalPrinter.SetUSBDevice(device, endpoint)
		resolve.Invoke(js.ValueOf(true))
	})
}

func setSerialPort(this js.Value, args []js.Value) any {
	if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
		globalPrinter.SetSerialPort(args[0])
		return js.ValueOf(true)
	}
	return js.ValueOf(false)
}

func setUSBDevice(this js.Value, args []js.Value) any {
	if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
		endpoint := 1
		if len(args) > 1 && !args[1].IsUndefined() && args[1].Type() == js.TypeNumber {
			endpoint = args[1].Int()
		}
		globalPrinter.SetUSBDevice(args[0], endpoint)
		return js.ValueOf(true)
	}
	return js.ValueOf(false)
}

func printTicket(this js.Value, args []js.Value) any {
	return newPromise(func(resolve, reject js.Value) {
		if len(args) == 0 {
			reject.Invoke("ticket data payload is required")
			return
		}

		data, err := parseTicketData(args[0])
		if err != nil {
			reject.Invoke(fmt.Sprintf("invalid ticket payload: %v", err))
			return
		}

		rawBytes := renderer.RenderTextTicket(data)
		n, err := globalPrinter.Write(rawBytes)
		if err != nil {
			reject.Invoke(fmt.Sprintf("print failed: %v", err))
			return
		}

		resolve.Invoke(js.ValueOf(n))
	})
}

func printTicketImage(this js.Value, args []js.Value) any {
	return newPromise(func(resolve, reject js.Value) {
		if len(args) == 0 {
			reject.Invoke("ticket data payload is required")
			return
		}

		widthDots := 576
		if len(args) > 1 && !args[1].IsUndefined() && args[1].Type() == js.TypeNumber {
			widthDots = args[1].Int()
		}

		data, err := parseTicketData(args[0])
		if err != nil {
			reject.Invoke(fmt.Sprintf("invalid ticket payload: %v", err))
			return
		}

		rawBytes := renderer.RenderImageTicket(data, widthDots)
		n, err := globalPrinter.Write(rawBytes)
		if err != nil {
			reject.Invoke(fmt.Sprintf("print image ticket failed: %v", err))
			return
		}

		resolve.Invoke(js.ValueOf(n))
	})
}

func printImageBase64(this js.Value, args []js.Value) any {
	return newPromise(func(resolve, reject js.Value) {
		if len(args) == 0 || args[0].Type() != js.TypeString {
			reject.Invoke("base64 image string is required")
			return
		}

		b64Str := args[0].String()
		// Strip Data URL prefix if present (e.g. data:image/png;base64,...)
		if idx := strings.Index(b64Str, ","); idx != -1 {
			b64Str = b64Str[idx+1:]
		}

		imgBytes, err := base64.StdEncoding.DecodeString(b64Str)
		if err != nil {
			reject.Invoke(fmt.Sprintf("failed to decode base64 image: %v", err))
			return
		}

		rawBytes, err := renderer.RenderRawImage(imgBytes)
		if err != nil {
			reject.Invoke(fmt.Sprintf("failed to render raster image: %v", err))
			return
		}

		n, err := globalPrinter.Write(rawBytes)
		if err != nil {
			reject.Invoke(fmt.Sprintf("print image failed: %v", err))
			return
		}

		resolve.Invoke(js.ValueOf(n))
	})
}

func printRawBytes(this js.Value, args []js.Value) any {
	return newPromise(func(resolve, reject js.Value) {
		if len(args) == 0 {
			reject.Invoke("Uint8Array data is required")
			return
		}

		uint8Arr := args[0]
		length := uint8Arr.Get("length").Int()
		data := make([]byte, length)
		js.CopyBytesToGo(data, uint8Arr)

		n, err := globalPrinter.Write(data)
		if err != nil {
			reject.Invoke(fmt.Sprintf("write raw bytes failed: %v", err))
			return
		}

		resolve.Invoke(js.ValueOf(n))
	})
}

func testPrint(this js.Value, args []js.Value) any {
	return newPromise(func(resolve, reject js.Value) {
		err := globalPrinter.TestPrint()
		if err != nil {
			reject.Invoke(fmt.Sprintf("test print failed: %v", err))
			return
		}
		resolve.Invoke(js.ValueOf(true))
	})
}

func closePrinter(this js.Value, args []js.Value) any {
	return newPromise(func(resolve, reject js.Value) {
		err := globalPrinter.Close()
		if err != nil {
			reject.Invoke(fmt.Sprintf("close failed: %v", err))
			return
		}
		resolve.Invoke(js.ValueOf(true))
	})
}

func getStatus(this js.Value, args []js.Value) any {
	res := js.Global().Get("Object").New()
	mode := globalPrinter.Mode()
	res.Set("mode", mode)
	res.Set("name", globalPrinter.Name())
	res.Set("connected", mode != "none")
	return res
}

func parseTicketData(val js.Value) (model.TicketData, error) {
	var data model.TicketData
	if val.Type() == js.TypeString {
		err := json.Unmarshal([]byte(val.String()), &data)
		return data, err
	} else if val.Type() == js.TypeObject {
		jsonCtor := js.Global().Get("JSON")
		jsonStr := jsonCtor.Call("stringify", val).String()
		err := json.Unmarshal([]byte(jsonStr), &data)
		return data, err
	}
	return data, errors.New("expected JSON string or JS Object")
}

func awaitPromise(promise js.Value) (js.Value, error) {
	ch := make(chan struct {
		val js.Value
		err error
	}, 1)

	thenFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		var v js.Value
		if len(args) > 0 {
			v = args[0]
		}
		ch <- struct {
			val js.Value
			err error
		}{val: v, err: nil}
		return nil
	})
	defer thenFn.Release()

	catchFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		errMsg := "promise rejected"
		if len(args) > 0 && !args[0].IsUndefined() {
			errMsg = args[0].String()
		}
		ch <- struct {
			val js.Value
			err error
		}{val: js.Null(), err: errors.New(errMsg)}
		return nil
	})
	defer catchFn.Release()

	promise.Call("then", thenFn).Call("catch", catchFn)

	res := <-ch
	return res.val, res.err
}
