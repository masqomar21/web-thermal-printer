const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const esbuild = require('esbuild');

const rootDir = path.resolve(__dirname, '..');
const distDir = path.join(rootDir, 'dist');
const wasmPath = path.join(distDir, 'printer.wasm');
const embeddedWasmTsPath = path.join(rootDir, 'src', 'embedded_wasm.ts');

console.log('🔨 Step 1: Building Go WebAssembly binary...');
execSync('bash build-wasm.sh', { cwd: rootDir, stdio: 'inherit' });

if (!fs.existsSync(wasmPath)) {
  console.error('❌ printer.wasm not found at ' + wasmPath);
  process.exit(1);
}

console.log('📦 Step 2: Embedding printer.wasm & wasm_exec.js into TypeScript source...');
const wasmBuffer = fs.readFileSync(wasmPath);
const wasmBase64 = wasmBuffer.toString('base64');
fs.writeFileSync(
  embeddedWasmTsPath,
  `// Auto-generated during build\nexport const EMBEDDED_WASM_BASE64 = "${wasmBase64}";\n`
);

const wasmExecJsPath = path.join(rootDir, 'web', 'wasm_exec.js');
if (fs.existsSync(wasmExecJsPath)) {
    let wasmExecJsCode = fs.readFileSync(wasmExecJsPath, 'utf-8');
  const polyfillSnippet = `\n\tif (!globalThis.crypto) {\n\t\tif (typeof require !== "undefined") {\n\t\t\ttry { globalThis.crypto = require("crypto").webcrypto; } catch (e) {}\n\t\t}\n\t}\n\tif (!globalThis.performance) {\n\t\tif (typeof require !== "undefined") {\n\t\t\ttry { globalThis.performance = require("perf_hooks").performance; } catch (e) {}\n\t\t}\n\t}\n`;
  wasmExecJsCode = wasmExecJsCode.replace('if (!globalThis.crypto) {', polyfillSnippet + '\tif (!globalThis.crypto) {');
  const wasmRuntimeTsPath = path.join(rootDir, 'src', 'wasm_exec_runtime.ts');
  fs.writeFileSync(
    wasmRuntimeTsPath,
    `// Auto-generated during build from official Go wasm_exec.js
export function ensureGoRuntime(): void {
  if (typeof (globalThis as any).Go !== 'undefined') {
    return;
  }
  const runCode = new Function(${JSON.stringify(wasmExecJsCode)});
  runCode();
}
`
  );
}

console.log('⚡ Step 3: Generating TypeScript declaration (.d.ts) files...');
try {
  execSync('npx tsc --emitDeclarationOnly', { cwd: rootDir, stdio: 'inherit' });
} catch (err) {
  console.warn('⚠️ tsc declaration warning:', err.message);
}

console.log('⚡ Step 4: Bundling ESM & CommonJS JavaScript modules with esbuild...');

// Build ESM (.mjs)
esbuild.buildSync({
  entryPoints: [path.join(rootDir, 'src', 'index.ts')],
  outfile: path.join(distDir, 'index.mjs'),
  bundle: true,
  format: 'esm',
  target: 'es2020',
  sourcemap: true,
});

// Build CommonJS (.cjs)
esbuild.buildSync({
  entryPoints: [path.join(rootDir, 'src', 'index.ts')],
  outfile: path.join(distDir, 'index.cjs'),
  bundle: true,
  format: 'cjs',
  target: 'es2020',
  sourcemap: true,
});

// Build Global IIFE (.global.js)
esbuild.buildSync({
  entryPoints: [path.join(rootDir, 'src', 'index.ts')],
  outfile: path.join(distDir, 'index.global.js'),
  bundle: true,
  format: 'iife',
  globalName: 'ThermalPrinterWASM',
  target: 'es2020',
  sourcemap: true,
});

console.log('✨ Build finished successfully! Outputs:');
console.log('   - ESM: dist/index.mjs');
console.log('   - CJS: dist/index.cjs');
console.log('   - IIFE: dist/index.global.js');
console.log('   - Types: dist/index.d.ts');

