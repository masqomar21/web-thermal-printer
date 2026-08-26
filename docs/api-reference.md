# API Reference - `thermal-printer-wasm`

Complete reference documentation for classes, methods, and TypeScript interfaces provided by `thermal-printer-wasm`.

---

## Class: `ThermalPrinterSDK`

The primary class for managing WebAssembly printer initialization, hardware connection, and printing execution.

```typescript
import { ThermalPrinterSDK } from 'thermal-printer-wasm';

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

#### `setUSBDevice(device: USBDevice, endpoint?: number): boolean`
Attaches an existing pre-opened `USBDevice` object directly.
- **Parameters**:
  - `device`: Native browser `USBDevice` object.
  - `endpoint` *(optional)*: Bulk Out Endpoint Number (default: `1`).
- **Returns**: `boolean`

#### `setSerialPort(port: SerialPort): boolean`
Attaches an existing pre-opened `SerialPort` object directly.
- **Parameters**:
  - `port`: Native browser `SerialPort` object.
- **Returns**: `boolean`

#### `print(payload: PrintPayload): Promise<number>`
Sends print job payload to printer.
- **Parameters**:
  - `payload`: `PrintTextOptions | PrintImageOptions | PrintRawOptions | string | Uint8Array`
- **Returns**: `Promise<number>` — Number of raw bytes written to printer hardware.

#### `testPrint(): Promise<boolean>`
Executes a standard built-in ESC/POS test page print.
- **Returns**: `Promise<boolean>`

#### `close(): Promise<boolean>`
Closes active printer hardware port and resets connection state.
- **Returns**: `Promise<boolean>`

#### `getStatus(): PrinterStatus`
Returns the current printer connection state synchronously.
- **Returns**: `PrinterStatus` object.

---

## TypeScript Interfaces

### `PrinterStatus`
```typescript
export interface PrinterStatus {
  /** Connection mode: 'none' | 'serial' | 'usb' */
  mode: string;
  /** Display name of the printer */
  name: string;
  /** Connection state boolean */
  connected: boolean;
}
```

### `PrintTextOptions`
```typescript
export interface PrintTextOptions {
  type: 'text';
  /** Text content string (supports multiline \n) */
  text: string;
  /** Alignment: 'left' | 'center' | 'right' (default: 'left') */
  align?: 'left' | 'center' | 'right';
  /** Bold font flag (default: false) */
  bold?: boolean;
  /** Feed lines count before paper cut (default: 3) */
  feed?: number;
  /** Execute ESC/POS paper cut (default: true) */
  cut?: boolean;
}
```

### `PrintImageOptions`
```typescript
export interface PrintImageOptions {
  type: 'image';
  /** Base64 encoded string or Data URI (e.g. data:image/png;base64,...) */
  base64: string;
}
```

### `PrintRawOptions`
```typescript
export interface PrintRawOptions {
  type: 'raw';
  /** Raw ESC/POS command bytes */
  bytes: Uint8Array;
}
```

### `PrintPayload`
```typescript
export type PrintPayload =
  | PrintTextOptions
  | PrintImageOptions
  | PrintRawOptions
  | string
  | Uint8Array;
```
