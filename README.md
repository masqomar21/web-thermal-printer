# 🖨️ thermal-printer-wasm

> **Universal WebAssembly Thermal Printer Gateway for Browser & Node.js with Full TypeScript Support** ⚡

`thermal-printer-wasm` is a lightweight, zero-configuration WebAssembly SDK that connects your web applications (React, Vue, Next.js, Angular, Svelte, or plain JS/TS) directly to **ESC/POS Thermal Receipt Printers** via **Web USB** or **Web Serial API**.

Zero background desktop software installation required on client machines.

---

## ✨ Features

- **⚡ Zero-Config Embedded WASM**: `printer.wasm` is Base64 embedded inside the npm bundle. Call `await printer.init()` without copying `.wasm` files or configuring bundlers (Vite, Next.js, Webpack, etc.).
- **📘 100% TypeScript Ready**: Includes strict TypeScript type declarations (`dist/index.d.ts`) out-of-the-box.
- **🔌 Native Browser Drivers**: Supports both **Web USB API** (`navigator.usb`) and **Web Serial API** (`navigator.serial`).
- **🎨 3 Print Modes**:
  1. **Custom Formatted Text**: Multiline text with alignment, bold font, line feeding, and paper cutting.
  2. **Graphic Image / Canvas**: Print HTML5 Canvas elements or Base64 images as high-precision thermal raster bitmaps.
  3. **Raw ESC/POS Bytes**: Pass raw `Uint8Array` command bytes directly to the printer.
- **📦 Isomorphic Dual Module**: Exported as **ESM** (`.mjs`), **CommonJS** (`.cjs`), and **Global IIFE** (`.global.js`).

---

## 📚 Documentation Index

- [🚀 Getting Started Guide](./docs/getting-started.md)
- [📘 API Reference & TypeScript Interfaces](./docs/api-reference.md)
- [🔌 Connection Guide (Web USB vs Web Serial)](./docs/connection-guide.md)
- [🎨 Printing Modes (Text, Canvas Image & Raw Bytes)](./docs/printing-modes.md)

---

## 📦 Installation

```bash
npm install thermal-printer-wasm
# or
yarn add thermal-printer-wasm
# or
pnpm add thermal-printer-wasm
```

---

## 🚀 Quick Start

### TypeScript / ES Modules (Vite, Next.js, React, Vue, Svelte)

```typescript
import { ThermalPrinterSDK, PrintPayload } from 'thermal-printer-wasm';

// 1. Instantiate SDK
const printer = new ThermalPrinterSDK();

async function printReceipt() {
  // 2. Initialize WASM engine (Zero-config embedded WASM)
  await printer.init();

  // 3. Connect via Web USB (Requires user gesture/button click in browser)
  await printer.connectUSB(1);

  // 4. Send Print Payload
  const payload: PrintPayload = {
    type: 'text',
    text: `TOKO UTAMA JAYA
Jl. Mawar No. 88, Jakarta
--------------------------------
No Nota : TRX-2026-001
Kasir   : Budi
--------------------------------
Kopi Latte      2x @20.000 = 40.000
Roti Toast      1x @15.000 = 15.000
--------------------------------
Total                      = 55.000
================================
Terima Kasih atas Kunjungan Anda!`,
    align: 'center',
    bold: true,
    cut: true,
    feed: 3,
  };

  const bytesSent = await printer.print(payload);
  console.log(`Print successful! Sent ${bytesSent} bytes.`);
}
```

### CommonJS (Node.js)

```javascript
const { ThermalPrinterSDK } = require('thermal-printer-wasm');

const printer = new ThermalPrinterSDK();
await printer.init();

console.log('Printer Status:', printer.getStatus());
```

### Plain HTML Script Tag (Browser ES Module)

```html
<script type="module">
  import { ThermalPrinterSDK } from 'https://unpkg.com/thermal-printer-wasm/dist/index.mjs';

  const printer = new ThermalPrinterSDK();
  await printer.init();

  document.getElementById('btnConnect').onclick = async () => {
    await printer.connectUSB(1);
    await printer.print('Hello Thermal Printer!\n');
  };
</script>
```

---

## 🎨 Printing Modes & Payloads

### 1. Formatted Text Mode (`type: 'text'`)

```typescript
await printer.print({
  type: 'text',
  text: 'STORE TITLE\nAddress Line\nItems...\n',
  align: 'center', // 'left' | 'center' | 'right' (default: 'left')
  bold: true,      // true | false (default: false)
  cut: true,       // execute paper cut (default: true)
  feed: 3,         // feed lines before cut (default: 3)
});
```

### 2. Canvas / Image Mode (`type: 'image'`)

Convert an HTML5 Canvas or Base64 PNG/JPEG image to thermal raster bytes:

```typescript
const canvas = document.getElementById('receiptCanvas') as HTMLCanvasElement;
const base64Image = canvas.toDataURL('image/png');

await printer.print({
  type: 'image',
  base64: base64Image,
});
```

### 3. Raw ESC/POS Bytes Mode (`type: 'raw'`)

Send raw `Uint8Array` command buffers directly:

```typescript
const rawBytes = new Uint8Array([27, 64, 27, 97, 49, 80, 82, 73, 78, 84, 10, 29, 86, 65, 48]);

await printer.print({
  type: 'raw',
  bytes: rawBytes,
});
```

---

## 🔌 Connection Guide: Web USB vs Web Serial

| Driver | Browser API | Target Devices | When to Use |
|---|---|---|---|
| **Web USB** | `navigator.usb` | Pure USB Thermal Printers | **Recommended** for standard USB thermal receipt printers. |
| **Web Serial** | `navigator.serial` | Virtual COM / Serial Ports (`/dev/cu.usbserial`, `COM3`) | For printers connected via RS232 Serial cable or Serial-to-USB converters (CH340, FTDI). |

### 🔒 Browser Security Policies & IP LAN Testing

Both Web USB and Web Serial APIs require a **Secure Context** (`HTTPS` or `localhost`).

- **Testing via IP LAN (HTTP)**: If testing from client devices over local IP (e.g. `http://192.168.1.100:8080`), Chrome will block Web Serial/Web USB by default.
- **Chrome Flag Workaround (Dev/Testing)**:
  1. In client Chrome, open `chrome://flags/#unsafely-treat-insecure-origin-as-secure`
  2. Enable the flag and enter your origin: `http://192.168.1.100:8080`
  3. Relaunch Chrome.

---

## 📘 TypeScript API Reference

### Class: `ThermalPrinterSDK`

#### `.init(wasmInput?: string | ArrayBuffer | Uint8Array): Promise<void>`
Initializes the WebAssembly runtime engine. By default, uses embedded Base64 WASM for zero-config initialization.

#### `.connectUSB(endpoint?: number): Promise<boolean>`
Prompts browser dialog to select and claim Web USB device. (Default endpoint: `1`).

#### `.connectSerial(baudRate?: number): Promise<boolean>`
Prompts browser dialog to select Web Serial port. (Default baudRate: `9600`).

#### `.print(payload: PrintPayload): Promise<number>`
Sends text, image, or raw byte payload to the connected printer. Returns number of bytes sent.

#### `.testPrint(): Promise<boolean>`
Prints a standard ESC/POS test page.

#### `.close(): Promise<boolean>`
Closes active printer connection and releases hardware locks.

#### `.getStatus(): PrinterStatus`
Returns current connection status:
```typescript
interface PrinterStatus {
  mode: 'none' | 'serial' | 'usb';
  name: string;
  connected: boolean;
}
```

---

## 🛠️ Development & Building

```bash
# Install dependencies
npm install

# Compile Go WASM & build npm package (ESM, CJS, IIFE, & .d.ts)
npm run build

# Run automated package import & WASM tests
npm test
```

---

## 📄 License

[MIT](LICENSE) © masqomar21
