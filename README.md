# Antrean Ticket Printer (Go Edition ⚡)

A fast, lightweight, single-binary application written in **Go** to print queue numbers using thermal printers. The application communicates with a server through **Socket.IO** and listens for print commands to generate and print queue tickets.

## Features

- **Dual Render Mode**: Supports both **Direct ESC/POS Text** (fast & lightweight) and **Native Go Raster Graphic Image** rendering.
- **Multiple Printer Connections**: Supports USB Thermal Printers (Auto-Detect), Network IP printers (TCP 9100), Serial/COM ports, and OS Print Spoolers.
- **Socket.IO Real-time Client**: Automatic reconnection, event topic customization, ACK response, and status update emission.
- **Zero Heavy Runtime**: No Node.js, Puppeteer, or Chrome required. Runs with under 20MB of RAM.

---

## Installation & Requirements

### Prerequisites
- [Go](https://go.dev/doc/install) 1.21 or later.
- A thermal printer (USB / Network / Serial / OS Spooler).

### Building from Source

```bash
git clone https://github.com/masqomar21/antrean-ticket-printer.git
cd antrean-ticket-printer
go mod tidy
go build -o ticket-printer ./cmd/ticket-printer
```

---

## Configuration

Edit the `.env` or `config.json` file to set server and printer configurations:

```env
SOCKET_URL="https://be.simpuskes.com"
TOPIC_PRINT_NOMOR_ANTREAN="antrean_print"
TOPIC_STATUS="status"

PRINTER_TYPE="usb"            # Options: "usb", "net", "serial", "system"
PRINTER_RENDER_MODE="text"    # Options: "text", "image", "both"
PRINTER_IP="192.168.1.200:9100"
PRINTER_SERIAL_PORT="COM3"
```

---

---

## Usage & Testing

### 1. Main Service (Socket.IO Print Service)
Start the production service:

```bash
go build -o ticket-printer ./cmd
./ticket-printer
```

or on Windows:

```cmd
run.bat
```

### 2. Printer Test Utility (Interactive CLI)
Test printer connectivity and print sample tickets without connecting to Socket.IO:

```bash
# Interactive test runner
./test.sh
```

or on Windows:

```cmd
test.bat
```

---

## Project Structure

```
antrean-ticket-printer/
├── cmd/
│   └── ticket-printer/       # Application main entrypoint
├── internal/
│   ├── config/               # Configuration loader (.env and config.json)
│   ├── model/                # Data models (TicketData, PrintStatusPayload)
│   ├── printer/              # ESC/POS command builder & drivers (USB, Net, Serial, Spooler)
│   ├── renderer/             # Text & Image ticket renderers
│   ├── service/              # Print service manager
│   └── socket/               # Socket.IO client manager
├── .env.example              # Environment file template
├── config.json.example       # JSON config template
├── go.mod                    # Go module definition
└── run.bat                   # Quick launcher script
```
