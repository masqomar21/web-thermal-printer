const assert = require('assert');

async function runTests() {
  console.log('🧪 Testing CommonJS import of web-thermal-printer...');
  const { ThermalPrinterSDK } = require('../dist/index.cjs');
  assert.ok(ThermalPrinterSDK, 'ThermalPrinterSDK should be exported in CJS');
  const printerCJS = new ThermalPrinterSDK();
  assert.ok(printerCJS, 'ThermalPrinterSDK instance created in CJS');
  console.log('✅ CJS export verified successfully');

  console.log('🧪 Testing ESM import of web-thermal-printer...');
  const esmModule = await import('../dist/index.mjs');
  assert.ok(esmModule.ThermalPrinterSDK, 'ThermalPrinterSDK should be exported in ESM');
  const printerESM = new esmModule.ThermalPrinterSDK();
  assert.ok(printerESM, 'ThermalPrinterSDK instance created in ESM');

  console.log('🧪 Testing printer.init() WebAssembly instantiation...');
  await printerESM.init();
  const status = printerESM.getStatus();
  assert.ok(status, 'getStatus() returned printer status');
  console.log('✅ WASM printer engine initialized successfully! Mode:', status.mode);
  console.log('🧪 Testing ESCPOSBuilder class export & chaining...');
  const { ESCPOSBuilder } = esmModule;
  assert.ok(ESCPOSBuilder, 'ESCPOSBuilder should be exported in ESM');
  const builder = new ESCPOSBuilder();
  builder
    .alignCenter()
    .setFontSize(2, 2)
    .setBold(true)
    .textLn('HEADER TEST')
    .setFontSize(1, 1)
    .setBold(false)
    .divider('-', 32)
    .tableLine('Item 1', '10.000', 32)
    .qrCode('https://example.com', 4)
    .barcodeCODE128('123456789')
    .cutPaper();

  const bytes = builder.toBytes();
  assert.ok(bytes instanceof Uint8Array, 'toBytes() should return Uint8Array');
  assert.ok(bytes.length > 50, 'Builder generated valid ESC/POS byte buffer');
  console.log(`✅ ESCPOSBuilder verified successfully! Byte length: ${bytes.length}`);

  console.log('🧪 Testing direct helper methods on ThermalPrinterSDK instance...');
  assert.strictEqual(typeof printerESM.printText, 'function', 'printText method exists');
  assert.strictEqual(typeof printerESM.printQRCode, 'function', 'printQRCode method exists');
  assert.strictEqual(typeof printerESM.printBarcode, 'function', 'printBarcode method exists');
  assert.strictEqual(typeof printerESM.printDivider, 'function', 'printDivider method exists');
  assert.strictEqual(typeof printerESM.printTableLine, 'function', 'printTableLine method exists');
  console.log('✅ All direct helper methods verified on ThermalPrinterSDK!');

  console.log('🎉 All package export tests passed cleanly!');
}

runTests().catch((err) => {
  console.error('❌ Test failed:', err);
  process.exit(1);
});
