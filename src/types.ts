/**
 * Connection status of the thermal printer
 */
export interface PrinterStatus {
  /** Connection mode: 'none' | 'serial' | 'usb' */
  mode: string;
  /** Device display name */
  name: string;
  /** Whether the printer is currently connected */
  connected: boolean;
}

/**
 * Text printing options
 */
export interface PrintTextOptions {
  /** Type identifier */
  type: 'text';
  /** Text content to print (supports newline \n) */
  text: string;
  /** Alignment: 'left' | 'center' | 'right' (default: 'left') */
  align?: 'left' | 'center' | 'right';
  /** Enable bold text (default: false) */
  bold?: boolean;
  /** Execute paper cut after printing (default: true) */
  cut?: boolean;
  /** Number of line feeds before cut (default: 3) */
  feed?: number;
}

/**
 * Graphic image printing options
 */
export interface PrintImageOptions {
  /** Type identifier */
  type: 'image';
  /** Base64 string or Data URI (e.g. data:image/png;base64,...) */
  base64: string;
}

/**
 * Raw ESC/POS bytes printing options
 */
export interface PrintRawOptions {
  /** Type identifier */
  type: 'raw';
  /** Raw Uint8Array containing ESC/POS command bytes */
  bytes: Uint8Array;
}

/**
 * Universal print payload (Union of Text, Image, Raw, or raw string/Uint8Array)
 */
export type PrintPayload =
  | PrintTextOptions
  | PrintImageOptions
  | PrintRawOptions
  | string
  | Uint8Array;

/**
 * Window extension interface for WASM global binding
 */
declare global {
  interface Window {
    ThermalPrinterWASM?: {
      connectSerial: (baudRate: number) => Promise<boolean>;
      connectUSB: (endpoint: number) => Promise<boolean>;
      setSerialPort: (port: any) => boolean;
      setUSBDevice: (device: any, endpoint?: number) => boolean;
      print: (payload: any) => Promise<number>;
      testPrint: () => Promise<boolean>;
      close: () => Promise<boolean>;
      getStatus: () => PrinterStatus;
    };
    Go?: any;
  }
}
