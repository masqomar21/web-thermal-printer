import {
  ThermalPrinterSDK,
  PrintPayload,
  PrintTextOptions,
  PrintImageOptions,
  PrintRawOptions,
  PrinterStatus,
} from '../src'; // Or 'thermal-printer-wasm' in published npm package

async function runTypeScriptExample() {
  console.log('🚀 Initializing ThermalPrinterSDK...');

  // 1. Create SDK instance
  const printer = new ThermalPrinterSDK();

  // 2. Initialize WASM engine (Zero-Config: Base64 embedded WASM & runtime auto-loaded)
  await printer.init();
  console.log('✅ SDK initialized successfully!');

  // 3. Check connection status
  const status: PrinterStatus = printer.getStatus();
  console.log('📊 Initial Status:', status);

  // 4. Example: Connect via Web USB (Bulk Out Endpoint 1)
  try {
    console.log('🔌 Requesting Web USB connection...');
    // Note: connectUSB() & connectSerial() require user gesture in browsers
    const connected: boolean = await printer.connectUSB(1);
    console.log('Connected to USB printer:', connected);
  } catch (err) {
    console.warn('USB connection skipped or failed:', (err as Error).message);
  }

  // 5. Example: Print Formatted Text Payload
  const textPayload: PrintTextOptions = {
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

  // 6. Example: Print Graphic Image Payload (Base64)
  const imagePayload: PrintImageOptions = {
    type: 'image',
    base64: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
  };

  // 7. Example: Print Raw ESC/POS Command Bytes
  const rawPayload: PrintRawOptions = {
    type: 'raw',
    bytes: new Uint8Array([27, 64, 27, 97, 49, 80, 82, 73, 78, 84, 10, 29, 86, 65, 48]),
  };

  console.log('📄 Payloads defined ready for printing:');
  console.log('- Text Payload:', textPayload.type);
  console.log('- Image Payload:', imagePayload.type);
  console.log('- Raw Payload:', rawPayload.bytes.length, 'bytes');

  // Example print call:
  // const bytesSent: number = await printer.print(textPayload);
  // console.log(`Printed ${bytesSent} bytes.`);
}

runTypeScriptExample().catch((err) => console.error('Error:', err));
