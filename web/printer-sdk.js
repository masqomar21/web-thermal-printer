/**
 * Thermal Printer WebAssembly SDK
 * Connects directly to ESC/POS thermal printers via Web Serial or Web USB API.
 * Pure Client-side Browser Gateway.
 */
class ThermalPrinterSDK {
  constructor() {
    this.initialized = false;
  }

  /**
   * Initialize Go WASM runtime & load printer.wasm binary
   * @param {string} wasmUrl Path to printer.wasm file (default: './printer.wasm')
   */
  async init(wasmUrl = './printer.wasm') {
    if (this.initialized) return;

    if (typeof Go === 'undefined') {
      throw new Error('wasm_exec.js must be loaded before initializing ThermalPrinterSDK');
    }

    const go = new Go();
    let result;

    if (WebAssembly.instantiateStreaming) {
      try {
        result = await WebAssembly.instantiateStreaming(fetch(wasmUrl), go.importObject);
      } catch (err) {
        console.warn('instantiateStreaming failed, falling back to arrayBuffer:', err);
        const resp = await fetch(wasmUrl);
        const bytes = await resp.arrayBuffer();
        result = await WebAssembly.instantiate(bytes, go.importObject);
      }
    } else {
      const resp = await fetch(wasmUrl);
      const bytes = await resp.arrayBuffer();
      result = await WebAssembly.instantiate(bytes, go.importObject);
    }

    go.run(result.instance);

    // Wait until window.ThermalPrinterWASM object is ready
    let attempts = 0;
    while (!window.ThermalPrinterWASM && attempts < 20) {
      await new Promise((r) => setTimeout(r, 50));
      attempts++;
    }

    if (!window.ThermalPrinterWASM) {
      throw new Error('Failed to initialize ThermalPrinterWASM binding');
    }

    this.initialized = true;
    console.log('✅ ThermalPrinterSDK initialized successfully');
  }

  /**
   * Request user permission and connect via Web Serial API (Virtual COM / USB Serial)
   * @param {number} baudRate Baud rate (default: 9600)
   */
  async connectSerial(baudRate = 9600) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.connectSerial(baudRate);
  }

  /**
   * Request user permission and connect via Web USB API
   * @param {number} endpoint Bulk Out Endpoint Number (default: 1)
   */
  async connectUSB(endpoint = 1) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.connectUSB(endpoint);
  }

  /**
   * Attach pre-opened SerialPort object directly
   * @param {SerialPort} port 
   */
  setSerialPort(port) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.setSerialPort(port);
  }

  /**
   * Attach pre-opened USBDevice object directly
   * @param {USBDevice} device 
   * @param {number} endpoint 
   */
  setUSBDevice(device, endpoint = 1) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.setUSBDevice(device, endpoint);
  }

  /**
   * Universal Print Method
   * Accepts:
   * 1. { type: "text", text: "Hello World", align: "center", bold: true, cut: true, feed: 3 }
   * 2. { type: "image", base64: "data:image/png;base64,..." }
   * 3. { type: "raw", bytes: Uint8Array }
   * 4. String or Uint8Array directly
   * @param {Object|string|Uint8Array} payload 
   */
  async print(payload) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.print(payload);
  }

  /**
   * Send test print
   */
  async testPrint() {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.testPrint();
  }

  /**
   * Close active printer connection
   */
  async close() {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.close();
  }

  /**
   * Get current printer connection status
   * @returns {{ mode: string, name: string, connected: boolean }}
   */
  getStatus() {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.getStatus();
  }

  ensureInitialized() {
    if (!this.initialized) {
      throw new Error('ThermalPrinterSDK is not initialized. Call await printer.init() first.');
    }
  }
}

// Export for module & script tag usage
if (typeof module !== 'undefined' && module.exports) {
  module.exports = ThermalPrinterSDK;
}
