import { ensureGoRuntime } from './wasm_exec_runtime';
import { EMBEDDED_WASM_BASE64 } from './embedded_wasm';
import type { PrinterStatus, PrintPayload } from './types';

declare const Buffer: any;

export * from './types';

/**
 * Universal WebAssembly Thermal Printer SDK
 */
export class ThermalPrinterSDK {
  private initialized = false;

  /**
   * Initialize Go WASM runtime & instantiate printer.wasm module.
   * By default, uses embedded Base64 WASM binary for zero-configuration initialization.
   * 
   * @param wasmInput Optional custom URL string, ArrayBuffer, or Uint8Array of printer.wasm
   */
  async init(wasmInput?: string | ArrayBuffer | Uint8Array): Promise<void> {
    if (this.initialized) return;

    ensureGoRuntime();

    const GoClass = (globalThis as any).Go;
    if (!GoClass) {
      throw new Error('Failed to load Go WebAssembly runtime');
    }

    const go = new GoClass();
    let instance: WebAssembly.Instance;

    if (wasmInput) {
      if (typeof wasmInput === 'string') {
        if (typeof WebAssembly.instantiateStreaming === 'function') {
          try {
            const response = await fetch(wasmInput);
            const res: any = await WebAssembly.instantiateStreaming(response, go.importObject);
            instance = res.instance || res;
          } catch (err) {
            const resp = await fetch(wasmInput);
            const bytes = await resp.arrayBuffer();
            const res: any = await WebAssembly.instantiate(bytes, go.importObject);
            instance = res.instance || res;
          }
        } else {
          const resp = await fetch(wasmInput);
          const bytes = await resp.arrayBuffer();
          const res: any = await WebAssembly.instantiate(bytes, go.importObject);
          instance = res.instance || res;
        }
      } else if (wasmInput instanceof ArrayBuffer || wasmInput instanceof Uint8Array) {
        const res: any = await WebAssembly.instantiate(wasmInput, go.importObject);
        instance = res.instance || res;
      } else {
        throw new Error('Invalid wasmInput argument type');
      }
    } else {
      // Zero-config: Use embedded Base64 WASM
      if (!EMBEDDED_WASM_BASE64) {
        throw new Error('Embedded WASM binary is empty. Run npm run build first.');
      }
      const bytes = this.base64ToUint8Array(EMBEDDED_WASM_BASE64);
      const res: any = await WebAssembly.instantiate(bytes, go.importObject);
      instance = res.instance || res;
    }

    go.run(instance);

    // Poll until window/globalThis.ThermalPrinterWASM is bound by WASM main()
    let attempts = 0;
    while (!(globalThis as any).ThermalPrinterWASM && attempts < 40) {
      await new Promise((r) => setTimeout(r, 50));
      attempts++;
    }

    if (!(globalThis as any).ThermalPrinterWASM) {
      throw new Error('Failed to bind ThermalPrinterWASM JS global object');
    }

    this.initialized = true;
  }

  /**
   * Request user permission and connect via Web Serial API
   * @param baudRate Baud rate (default: 9600)
   */
  async connectSerial(baudRate = 9600): Promise<boolean> {
    this.ensureInitialized();
    return (globalThis as any).ThermalPrinterWASM.connectSerial(baudRate);
  }

  /**
   * Request user permission and connect via Web USB API
   * @param endpoint Bulk Out Endpoint Number (default: 1)
   */
  async connectUSB(endpoint = 1): Promise<boolean> {
    this.ensureInitialized();
    return (globalThis as any).ThermalPrinterWASM.connectUSB(endpoint);
  }

  /**
   * Attach pre-opened SerialPort object directly
   * @param port SerialPort instance
   */
  setSerialPort(port: any): boolean {
    this.ensureInitialized();
    return (globalThis as any).ThermalPrinterWASM.setSerialPort(port);
  }

  /**
   * Attach pre-opened USBDevice object directly
   * @param device USBDevice instance
   * @param endpoint Bulk Out Endpoint Number (default: 1)
   */
  setUSBDevice(device: any, endpoint = 1): boolean {
    this.ensureInitialized();
    return (globalThis as any).ThermalPrinterWASM.setUSBDevice(device, endpoint);
  }

  /**
   * Universal Print Method
   * Accepts text formatting object, image base64, raw Uint8Array, or raw text string.
   * @param payload Print options or raw data
   */
  async print(payload: PrintPayload): Promise<number> {
    this.ensureInitialized();
    return (globalThis as any).ThermalPrinterWASM.print(payload);
  }

  /**
   * Execute test print
   */
  async testPrint(): Promise<boolean> {
    this.ensureInitialized();
    return (globalThis as any).ThermalPrinterWASM.testPrint();
  }

  /**
   * Close active printer connection
   */
  async close(): Promise<boolean> {
    this.ensureInitialized();
    return (globalThis as any).ThermalPrinterWASM.close();
  }

  /**
   * Get current printer connection status
   */
  getStatus(): PrinterStatus {
    this.ensureInitialized();
    return (globalThis as any).ThermalPrinterWASM.getStatus();
  }

  private ensureInitialized(): void {
    if (!this.initialized) {
      throw new Error('ThermalPrinterSDK is not initialized. Call await printer.init() first.');
    }
  }

  private base64ToUint8Array(base64: string): Uint8Array {
    if (typeof atob === 'function') {
      const binaryString = atob(base64);
      const len = binaryString.length;
      const bytes = new Uint8Array(len);
      for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }
      return bytes;
    } else if (typeof Buffer !== 'undefined') {
      return new Uint8Array(Buffer.from(base64, 'base64'));
    }
    throw new Error('No base64 decoder available in current environment');
  }
}

export default ThermalPrinterSDK;
