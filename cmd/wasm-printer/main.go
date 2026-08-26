//go:build js && wasm

package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"syscall/js"

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
	sdkObj.Set("print", js.FuncOf(printUniversal))
	sdkObj.Set("printText", js.FuncOf(printTextDirect))
	sdkObj.Set("printImage", js.FuncOf(printImageDirect))
	sdkObj.Set("printRaw", js.FuncOf(printRawDirect))
	sdkObj.Set("printQRCode", js.FuncOf(printQRCodeDirect))
	sdkObj.Set("printBarcode", js.FuncOf(printBarcodeDirect))
	sdkObj.Set("printDivider", js.FuncOf(printDividerDirect))
	sdkObj.Set("printTableLine", js.FuncOf(printTableLineDirect))
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

		portPromise := nav.Get("serial").Call("requestPort")
		port, err := awaitPromise(portPromise)
		if err != nil {
			reject.Invoke(fmt.Sprintf("Failed to select serial port: %v", err))
			return
		}

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

// printUniversal is the single universal print API for JS
// Accepts:
// 1. { type: "text", text: "..." }
// 2. { type: "image", base64: "data:image/png;base64,..." }
// 3. { type: "raw", bytes: Uint8Array }
func printUniversal(this js.Value, args []js.Value) any {
	return newPromise(func(resolve, reject js.Value) {
		if len(args) == 0 {
			reject.Invoke("print payload is required")
			return
		}

		arg := args[0]
		pType := "text"

		if arg.Type() == js.TypeObject && !arg.Get("type").IsUndefined() {
			pType = strings.ToLower(arg.Get("type").String())
		} else if arg.Type() == js.TypeString {
			pType = "text"
		} else if arg.Type() == js.TypeObject && !arg.Get("byteLength").IsUndefined() {
			pType = "raw"
		}

		switch pType {
		case "image":
			b64Str := ""
			if arg.Type() == js.TypeString {
				b64Str = arg.String()
			} else if !arg.Get("base64").IsUndefined() {
				b64Str = arg.Get("base64").String()
			}
			if b64Str == "" {
				reject.Invoke("base64 image payload is required")
				return
			}
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

		case "raw":
			var uint8Arr js.Value
			if !arg.Get("bytes").IsUndefined() {
				uint8Arr = arg.Get("bytes")
			} else {
				uint8Arr = arg
			}
			length := uint8Arr.Get("length").Int()
			data := make([]byte, length)
			js.CopyBytesToGo(data, uint8Arr)
			n, err := globalPrinter.Write(data)
			if err != nil {
				reject.Invoke(fmt.Sprintf("write raw bytes failed: %v", err))
				return
			}
			resolve.Invoke(js.ValueOf(n))

		case "qrcode":
			content := ""
			size := 4
			align := "center"
			cut := false

			if !arg.Get("content").IsUndefined() {
				content = arg.Get("content").String()
			}
			if !arg.Get("size").IsUndefined() {
				size = arg.Get("size").Int()
			}
			if !arg.Get("align").IsUndefined() {
				align = strings.ToLower(arg.Get("align").String())
			}
			if !arg.Get("cut").IsUndefined() {
				cut = arg.Get("cut").Bool()
			}

			if content == "" {
				reject.Invoke("qrcode content payload is required")
				return
			}

			b := printer.NewESCPOSBuilder()
			switch align {
			case "center":
				b.AlignCenter()
			case "right":
				b.AlignRight()
			default:
				b.AlignLeft()
			}
			b.QRCode(content, size)
			if cut {
				b.CutPaper()
			}

			n, err := globalPrinter.Write(b.Bytes())
			if err != nil {
				reject.Invoke(fmt.Sprintf("print qrcode failed: %v", err))
				return
			}
			resolve.Invoke(js.ValueOf(n))

		case "barcode":
			content := ""
			align := "center"
			cut := false

			if !arg.Get("content").IsUndefined() {
				content = arg.Get("content").String()
			}
			if !arg.Get("align").IsUndefined() {
				align = strings.ToLower(arg.Get("align").String())
			}
			if !arg.Get("cut").IsUndefined() {
				cut = arg.Get("cut").Bool()
			}

			if content == "" {
				reject.Invoke("barcode content payload is required")
				return
			}

			b := printer.NewESCPOSBuilder()
			switch align {
			case "center":
				b.AlignCenter()
			case "right":
				b.AlignRight()
			default:
				b.AlignLeft()
			}
			b.BarcodeCODE128(content)
			if cut {
				b.CutPaper()
			}

			n, err := globalPrinter.Write(b.Bytes())
			if err != nil {
				reject.Invoke(fmt.Sprintf("print barcode failed: %v", err))
				return
			}
			resolve.Invoke(js.ValueOf(n))

		case "divider":
			char := "-"
			width := 32

			if !arg.Get("char").IsUndefined() {
				char = arg.Get("char").String()
			}
			if !arg.Get("width").IsUndefined() {
				width = arg.Get("width").Int()
			}

			b := printer.NewESCPOSBuilder()
			b.Divider(char, width)

			n, err := globalPrinter.Write(b.Bytes())
			if err != nil {
				reject.Invoke(fmt.Sprintf("print divider failed: %v", err))
				return
			}
			resolve.Invoke(js.ValueOf(n))

		case "table":
			left := ""
			right := ""
			width := 32

			if !arg.Get("left").IsUndefined() {
				left = arg.Get("left").String()
			}
			if !arg.Get("right").IsUndefined() {
				right = arg.Get("right").String()
			}
			if !arg.Get("width").IsUndefined() {
				width = arg.Get("width").Int()
			}

			b := printer.NewESCPOSBuilder()
			b.TableLine(left, right, width)

			n, err := globalPrinter.Write(b.Bytes())
			if err != nil {
				reject.Invoke(fmt.Sprintf("print table failed: %v", err))
				return
			}
			resolve.Invoke(js.ValueOf(n))

		default: // "text"
			textStr := ""
			align := "left"
			bold := false
			cut := true
			feed := 3

			if arg.Type() == js.TypeString {
				textStr = arg.String()
			} else if arg.Type() == js.TypeObject {
				if !arg.Get("text").IsUndefined() {
					textStr = arg.Get("text").String()
				}
				if !arg.Get("align").IsUndefined() {
					align = strings.ToLower(arg.Get("align").String())
				}
				if !arg.Get("bold").IsUndefined() {
					bold = arg.Get("bold").Bool()
				}
				if !arg.Get("cut").IsUndefined() {
					cut = arg.Get("cut").Bool()
				}
				if !arg.Get("feed").IsUndefined() {
					feed = arg.Get("feed").Int()
				}
			}

			if textStr == "" {
				reject.Invoke("text payload is required")
				return
			}

			b := printer.NewESCPOSBuilder()
			switch align {
			case "center":
				b.AlignCenter()
			case "right":
				b.AlignRight()
			default:
				b.AlignLeft()
			}
			b.SetBold(bold)
			b.Text(textStr)
			if !strings.HasSuffix(textStr, "\n") {
				b.NewLine(1)
			}
			if feed > 0 {
				b.NewLine(feed)
			}
			if cut {
				b.CutPaper()
			}

			n, err := globalPrinter.Write(b.Bytes())
			if err != nil {
				reject.Invoke(fmt.Sprintf("print text failed: %v", err))
				return
			}
			resolve.Invoke(js.ValueOf(n))
		}
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

func printTextDirect(this js.Value, args []js.Value) any {
	return printUniversal(this, args)
}

func printImageDirect(this js.Value, args []js.Value) any {
	if len(args) > 0 && args[0].Type() == js.TypeString {
		obj := js.Global().Get("Object").New()
		obj.Set("type", "image")
		obj.Set("base64", args[0])
		return printUniversal(this, []js.Value{obj})
	}
	return printUniversal(this, args)
}

func printRawDirect(this js.Value, args []js.Value) any {
	if len(args) > 0 {
		obj := js.Global().Get("Object").New()
		obj.Set("type", "raw")
		obj.Set("bytes", args[0])
		return printUniversal(this, []js.Value{obj})
	}
	return printUniversal(this, args)
}

func printQRCodeDirect(this js.Value, args []js.Value) any {
	if len(args) > 0 && args[0].Type() == js.TypeString {
		obj := js.Global().Get("Object").New()
		obj.Set("type", "qrcode")
		obj.Set("content", args[0])
		if len(args) > 1 {
			obj.Set("size", args[1])
		}
		if len(args) > 2 {
			obj.Set("align", args[2])
		}
		if len(args) > 3 {
			obj.Set("cut", args[3])
		}
		return printUniversal(this, []js.Value{obj})
	}
	return printUniversal(this, args)
}

func printBarcodeDirect(this js.Value, args []js.Value) any {
	if len(args) > 0 && args[0].Type() == js.TypeString {
		obj := js.Global().Get("Object").New()
		obj.Set("type", "barcode")
		obj.Set("content", args[0])
		if len(args) > 1 {
			obj.Set("align", args[1])
		}
		if len(args) > 2 {
			obj.Set("cut", args[2])
		}
		return printUniversal(this, []js.Value{obj})
	}
	return printUniversal(this, args)
}

func printDividerDirect(this js.Value, args []js.Value) any {
	obj := js.Global().Get("Object").New()
	obj.Set("type", "divider")
	if len(args) > 0 {
		obj.Set("char", args[0])
	}
	if len(args) > 1 {
		obj.Set("width", args[1])
	}
	return printUniversal(this, []js.Value{obj})
}

func printTableLineDirect(this js.Value, args []js.Value) any {
	obj := js.Global().Get("Object").New()
	obj.Set("type", "table")
	if len(args) > 0 {
		obj.Set("left", args[0])
	}
	if len(args) > 1 {
		obj.Set("right", args[1])
	}
	if len(args) > 2 {
		obj.Set("width", args[2])
	}
	return printUniversal(this, []js.Value{obj})
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
