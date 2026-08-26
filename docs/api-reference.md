# API Reference - `web-thermal-printer`

Complete reference documentation for classes, methods, and TypeScript interfaces provided by `web-thermal-printer`.

---

## Class: `ThermalPrinterSDK`

The primary class for managing WebAssembly printer initialization, hardware connection, and printing execution.

```typescript
import { ThermalPrinterSDK, ESCPOSBuilder } from 'web-thermal-printer';

const printer = new ThermalPrinterSDK();
```

---

### Methods

#### `init(wasmInput?: string | ArrayBuffer | Uint8Array): Promise<void>`
Initializes the Go WebAssembly runtime and binds global JS functions.
- **Parameters**:
  - `wasmInput` *(optional)*: Custom URL string, `ArrayBuffer`, or `Uint8Array` of `printer.wasm`. If omitted, uses embedded Base64 WASM binary.
- **Returns**: `Promise<void>`
- **Throws**: `Error` if WASM instantiation fails.

#### `createBuilder(): ESCPOSBuilder`
Creates a new chainable `ESCPOSBuilder` instance for constructing raw ESC/POS commands.

#### `connectUSB(endpoint?: number): Promise<boolean>`
Opens browser permission prompt to connect to a Web USB thermal printer. Must be called inside a user gesture handler (e.g. `click` listener).
- **Parameters**:
  - `endpoint` *(optional)*: Bulk Out Endpoint Number (default: `1`).
- **Returns**: `Promise<boolean>` — Resolves `true` when connected.

#### `connectSerial(baudRate?: number): Promise<boolean>`
Opens browser permission prompt to connect to a Web Serial Virtual COM port. Must be called inside a user gesture handler.
- **Parameters**:
  - `baudRate` *(optional)*: Baud rate in bps (default: `9600`).
- **Returns**: `Promise<boolean>` — Resolves `true` when connected.

#### `print(payload: PrintPayload | ESCPOSBuilder): Promise<number>`
Sends print job payload or builder to printer.
- **Parameters**:
  - `payload`: `PrintTextOptions | PrintImageOptions | PrintRawOptions | PrintQRCodeOptions | PrintBarcodeOptions | ESCPOSBuilder | string | Uint8Array`
- **Returns**: `Promise<number>` — Number of raw bytes written to printer hardware.

#### `printText(text: string, options?: PrintTextOptions): Promise<number>`
Prints formatted text directly.

#### `printImage(base64: string): Promise<number>`
Prints raster image from Base64 or Data URI directly.

#### `printRaw(bytes: Uint8Array): Promise<number>`
Prints raw ESC/POS command bytes directly.

#### `printQRCode(content: string, size?: number, align?: 'left'|'center'|'right', cut?: boolean): Promise<number>`
Prints hardware QR Code directly (`GS ( k`).

#### `printBarcode(content: string, align?: 'left'|'center'|'right', cut?: boolean): Promise<number>`
Prints CODE128 barcode directly (`GS k 73`).

#### `printDivider(char?: string, width?: number): Promise<number>`
Prints horizontal line divider directly.

#### `printTableLine(left: string, right: string, width?: number): Promise<number>`
Prints 2-column table line directly.

---

## Class: `ESCPOSBuilder`

Chainable command builder for constructing ESC/POS command byte buffers directly in JavaScript / TypeScript.

```typescript
const builder = new ESCPOSBuilder();
builder
  .alignCenter()
  .setFontSize(2, 2)
  .setBold(true)
  .textLn("HEADER TITLE")
  .setFontSize(1, 1)
  .setBold(false)
  .divider("-", 32)
  .tableLine("Item 1", "20.000", 32)
  .qrCode("https://example.com", 4)
  .barcodeCODE128("12345678")
  .cutPaper();

await printer.print(builder);
```

### Methods
- `init()`: Initialize printer (`ESC @`).
- `alignLeft()`, `alignCenter()`, `alignRight()`: Set text alignment (`ESC a n`).
- `setFontSize(widthMulti, heightMulti)`: Set font magnification 1-8 (`GS ! n`).
- `setBold(enable)`: Enable or disable bold mode (`ESC E n`).
- `text(str)`, `textLn(str)`: Append text string.
- `newLine(count)`: Append empty lines (`\n`).
- `cutPaper()`: Execute paper cut (`GS V 65 0`).
- `divider(char?, lineLength?)`: Append horizontal line divider.
- `tableLine(left, right, totalWidth?)`: Append 2-column table line with left & right alignment.
- `qrCode(content, moduleSize?)`: Append hardware QR Code (`GS ( k`).
- `barcodeCODE128(content)`: Append CODE128 barcode (`GS k 73`).
- `toBytes()`: Returns compiled `Uint8Array`.
