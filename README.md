# Antrean Ticket Printer (Go Edition ⚡)

A fast, lightweight, single-binary application written in **Go** to print queue numbers using thermal printers. The application communicates with a server through **Socket.IO** and listens for print commands to generate and print queue tickets.

## Features

- **WebAssembly (WASM) Browser Gateway**: Compile to WebAssembly and print directly from public web SaaS applications using **Web Serial API** or **Web USB API** without client software installation!
- **Dual Render Mode**: Supports both **Direct ESC/POS Text** (fast & lightweight) and **Native Go Raster Graphic Image** rendering.
- **Multiple Printer Connections**: Supports USB Thermal Printers (Auto-Detect), Network IP printers (TCP 9100), Serial/COM ports, OS Print Spoolers, and Web Serial/USB (in-browser WASM).
- **Socket.IO Real-time Client**: Automatic reconnection, event topic customization, ACK response, and status update emission.
- **Zero Heavy Runtime**: No Node.js, Puppeteer, or Chrome required. Runs with under 20MB of RAM.

---

## Installation & Requirements

### Prerequisites
- [Go](https://go.dev/doc/install) 1.21 or later.
- A thermal printer (USB / Network / Serial / OS Spooler / Web Serial).

### Building Native Desktop App
```bash
go build -o ticket-printer ./cmd/main.go
```

### Building WASM Web Gateway (SaaS Public Browser Mode)
```bash
# Build Go WebAssembly binary
./build-wasm.sh

# Or on Windows:
build-wasm.bat
```

This compiles `web/printer.wasm` which can be served by any static file server alongside `web/wasm_exec.js`, `web/printer-sdk.js`, and `web/index.html`.

---

## Usage & Testing

### 1. WebAssembly SaaS Browser Demo (No Local Agent Required)
Serve the `web/` directory using any HTTP server:

```bash
npx serve web
# or
python3 -m http.server 8080 -d web
```
Open `http://localhost:8080` in Google Chrome or Microsoft Edge to test Web Serial/USB printing directly from the browser!

```javascript
// JS SDK Integration Example:
const printer = new ThermalPrinterSDK();
await printer.init('./printer.wasm');

// Request Web Serial permission & connect
await printer.connectSerial(9600);

// Print structured ticket
await printer.printTicket({
  instansi: "PUSKESMAS MAJU SEHAT",
  loket: "LOKET 1",
  nomor_antrean: "A-001",
  tanggal: "24/08/2026",
  jam: "10:30"
});
```

### 2. Main Service (Desktop Socket.IO Background Service)
Start the background daemon service:

```bash
go build -o ticket-printer ./cmd/main.go
./ticket-printer
```

or on Windows:

```cmd
run.bat
```

### 3. Printer Test Utility (Interactive CLI)
Test printer connectivity and print sample tickets without connecting to Socket.IO:

```bash
./test.sh
```

---

## Project Structure

```
antrean-ticket-printer/
├── cmd/
│   ├── main.go               # Desktop Socket.IO background service entrypoint
│   └── wasm-printer/         # WebAssembly entrypoint (syscall/js binding)
├── internal/
│   ├── config/               # Configuration loader (.env and config.json)
│   ├── model/                # Data models (TicketData, PrintStatusPayload)
│   ├── printer/              # ESC/POS command builder & drivers (USB, Net, Serial, Spooler, WASM)
│   ├── renderer/             # Text, Graphic Raster, and Raw Image ticket renderers
│   ├── service/              # Print service manager
│   └── socket/               # Socket.IO client manager
├── web/
│   ├── index.html            # Interactive SaaS browser printer studio demo
│   ├── printer-sdk.js        # JavaScript Promise-based SDK wrapper
│   ├── printer.wasm          # Compiled Go WASM binary engine
│   └── wasm_exec.js          # Go WASM browser runtime glue
├── build-wasm.sh             # Linux/Mac WASM build script
├── build-wasm.bat            # Windows WASM build script
├── go.mod                    # Go module definition
└── run.bat                   # Quick launcher script
```

