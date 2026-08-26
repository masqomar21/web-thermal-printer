# Getting Started with `thermal-printer-wasm`

Welcome to `thermal-printer-wasm`! This guide will help you install, initialize, and run your first thermal print job using WebAssembly in modern web applications.

---

## 1. Installation

Install the package into your project using your preferred package manager:

```bash
# npm
npm install thermal-printer-wasm

# yarn
yarn add thermal-printer-wasm

# pnpm
pnpm add thermal-printer-wasm
```

---

## 2. Zero-Config Initialization

`thermal-printer-wasm` embeds the compiled Go WebAssembly binary (`printer.wasm`) directly inside the npm JavaScript bundle as a Base64 string. 

This means **you do NOT need to**:
- Copy `.wasm` files to your project's `public/` directory.
- Configure WebAssembly file loaders in Vite, Next.js, Webpack, or Rollup.
- Manually include `wasm_exec.js` script tags in your HTML.

Simply instantiate `ThermalPrinterSDK` and call `await printer.init()`:

```typescript
import { ThermalPrinterSDK } from 'thermal-printer-wasm';

const printer = new ThermalPrinterSDK();

async function initPrinter() {
  await printer.init();
  console.log('✅ Printer engine initialized and ready!');
}
```

---

## 3. Basic Printing Example

Here is a complete example of connecting via Web USB and printing a receipt:

```typescript
import { ThermalPrinterSDK } from 'thermal-printer-wasm';

const printer = new ThermalPrinterSDK();

document.getElementById('btnPrint')?.addEventListener('click', async () => {
  try {
    // 1. Initialize WASM engine
    await printer.init();

    // 2. Request user permission & connect via Web USB
    // (Must be triggered by a user gesture like a button click)
    await printer.connectUSB(1);

    // 3. Print receipt
    const bytesSent = await printer.print({
      type: 'text',
      text: `TOKO UTAMA JAYA
--------------------------------
Kopi Latte      2x @20.000 = 40.000
Roti Toast      1x @15.000 = 15.000
--------------------------------
Total                      = 55.000
================================
Terima Kasih!`,
      align: 'center',
      bold: true,
      cut: true,
      feed: 3,
    });

    console.log(`Print successful! Sent ${bytesSent} bytes to printer.`);
  } catch (err) {
    console.error('Print failed:', err);
  }
});
```

---

## 4. Next Steps

- Check out [Connection Guide](./connection-guide.md) to learn about Web USB vs Web Serial and browser security policies.
- Check out [Printing Modes](./printing-modes.md) for custom text formatting, canvas image rendering, and raw byte sending.
- View [API Reference](./api-reference.md) for full method & TypeScript interface specifications.
