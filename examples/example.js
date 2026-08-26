const { ThermalPrinterSDK } = require('../dist/index.cjs'); // Or require('thermal-printer-wasm')

async function runJavaScriptExample() {
  console.log('🚀 Initializing ThermalPrinterSDK (JavaScript CommonJS)...');

  // 1. Create SDK instance
  const printer = new ThermalPrinterSDK();

  // 2. Initialize WASM engine (Zero-Config)
  await printer.init();
  console.log('✅ SDK initialized successfully!');

  // 3. Check status
  const status = printer.getStatus();
  console.log('📊 Current Status:', status);

  // 4. Print formatted text example
  const textPayload = {
    type: 'text',
    text: 'TOKO UTAMA JAYA\nJl. Mawar No. 88\nTotal: Rp 50.000\n',
    align: 'center',
    bold: true,
    cut: true,
    feed: 3,
  };

  console.log('📄 JavaScript Example Payload Ready:', textPayload);
}

runJavaScriptExample().catch((err) => console.error('Error:', err));
