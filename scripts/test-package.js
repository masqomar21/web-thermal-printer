const assert = require('assert');

async function runTests() {
  console.log('🧪 Testing CommonJS import of thermal-printer-wasm...');
  const { ThermalPrinterSDK } = require('../dist/index.cjs');
  assert.ok(ThermalPrinterSDK, 'ThermalPrinterSDK should be exported in CJS');
  const printerCJS = new ThermalPrinterSDK();
  assert.ok(printerCJS, 'ThermalPrinterSDK instance created in CJS');
  console.log('✅ CJS export verified successfully');

  console.log('🧪 Testing ESM import of thermal-printer-wasm...');
  const esmModule = await import('../dist/index.mjs');
  assert.ok(esmModule.ThermalPrinterSDK, 'ThermalPrinterSDK should be exported in ESM');
  const printerESM = new esmModule.ThermalPrinterSDK();
  assert.ok(printerESM, 'ThermalPrinterSDK instance created in ESM');

  console.log('🧪 Testing printer.init() WebAssembly instantiation...');
  await printerESM.init();
  const status = printerESM.getStatus();
  assert.ok(status, 'getStatus() returned printer status');
  console.log('✅ WASM printer engine initialized successfully! Mode:', status.mode);
  console.log('✅ ESM export verified successfully');

  console.log('🎉 All package export tests passed cleanly!');
}

runTests().catch((err) => {
  console.error('❌ Test failed:', err);
  process.exit(1);
});
