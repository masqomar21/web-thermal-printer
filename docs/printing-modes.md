# Printing Modes Guide

`thermal-printer-wasm` supports three primary printing modes: **Formatted Text Mode**, **Canvas Graphic Image Mode**, and **Raw ESC/POS Bytes Mode**.

---

## 1. Formatted Text Mode (`type: 'text'`)

Formatted Text Mode sends text directly using hardware font ESC/POS text commands. It is extremely fast, lightweight, and produces crisp text on all thermal printers.

```typescript
await printer.print({
  type: 'text',
  text: `TOKO MAJU BERSAMA
Jl. Merdeka No. 123
--------------------------------
Kopi Latte      2x @20.000 = 40.000
--------------------------------
Total                      = 40.000`,
  align: 'center', // 'left' | 'center' | 'right'
  bold: true,      // true | false
  cut: true,       // true | false
  feed: 3,         // feed line count before cut
});
```

### Options:
- **`align`**: Sets paragraph alignment (`ESC a 0`, `ESC a 1`, `ESC a 2`).
- **`bold`**: Enables bold font mode (`ESC E 1`).
- **`feed`**: Feeds `N` newline characters before paper cut.
- **`cut`**: Executes full paper cut command (`GS V 65 0`).

---

## 2. Canvas Graphic Image Mode (`type: 'image'`)

Canvas Image Mode converts any image (PNG, JPEG, Data URI) or HTML5 Canvas element into a thermal raster bitmap (`GS v 0`).

This mode allows complete freedom to print custom graphics, logos, custom fonts, barcodes, QR codes, tables, or complex ticket layouts designed on HTML5 Canvas.

```typescript
// 1. Draw receipt on HTML5 Canvas
const canvas = document.getElementById('receiptCanvas') as HTMLCanvasElement;
const ctx = canvas.getContext('2d')!;

ctx.fillStyle = '#ffffff';
ctx.fillRect(0, 0, canvas.width, canvas.height);

ctx.fillStyle = '#000000';
ctx.font = 'bold 20px sans-serif';
ctx.textAlign = 'center';
ctx.fillText('CUSTOM CANVAS RECEIPT', canvas.width / 2, 40);

// 2. Get Data URI
const base64DataUri = canvas.toDataURL('image/png');

// 3. Print raster bitmap
await printer.print({
  type: 'image',
  base64: base64DataUri,
});
```

---

## 3. Raw Bytes Mode (`type: 'raw'`)

Raw Bytes Mode allows you to send pre-rendered or custom ESC/POS binary buffers directly to the printer hardware.

Use this mode if you are using an external ESC/POS builder library or generating raw hardware bytes:

```typescript
// ESC @ (Init) + ESC a 1 (Align Center) + "PRINT OK" + \n + GS V 65 0 (Cut)
const escposBytes = new Uint8Array([
  0x1b, 0x40, 
  0x1b, 0x61, 0x01, 
  0x50, 0x52, 0x49, 0x4e, 0x54, 0x20, 0x4f, 0x4b, 0x0a, 
  0x0a, 0x0a, 
  0x1d, 0x56, 0x41, 0x00
]);

await printer.print({
  type: 'raw',
  bytes: escposBytes,
});
```
