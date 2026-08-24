/**
 * Thermal Printer WebAssembly SDK
 * Connects directly to ESC/POS thermal printers via Web Serial or Web USB API.
 * No background desktop installation required.
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
   * Print structured Queue Ticket in Text ESC/POS Mode
   * @param {Object|string} ticketData Ticket object: { instansi, loket, nomor_antrean, tanggal, jam, layanan }
   */
  async printTicket(ticketData) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.printTicket(ticketData);
  }

  /**
   * Print structured Queue Ticket in Graphic Raster ESC/POS Mode
   * @param {Object|string} ticketData 
   * @param {number} widthDots Printer width in dots (default 576 for 80mm, 384 for 58mm)
   */
  async printTicketImage(ticketData, widthDots = 576) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.printTicketImage(ticketData, widthDots);
  }

  /**
   * Print raw image from Base64 string or Data URI (PNG / JPEG)
   * @param {string} base64Str Base64 encoded image string or Data URL
   */
  async printImageBase64(base64Str) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.printImageBase64(base64Str);
  }

  /**
   * Print raw byte array of ESC/POS commands
   * @param {Uint8Array} uint8Array Raw byte buffer
   */
  async printRawBytes(uint8Array) {
    this.ensureInitialized();
    return window.ThermalPrinterWASM.printRawBytes(uint8Array);
  }

  /**
   * Send test print ticket
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
