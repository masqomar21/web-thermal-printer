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

console.log('📦 Step 2: Embedding printer.wasm as Base64 into src/embedded_wasm.ts...');
const wasmBuffer = fs.readFileSync(wasmPath);
const wasmBase64 = wasmBuffer.toString('base64');
fs.writeFileSync(
  embeddedWasmTsPath,
  `// Auto-generated during build\nexport const EMBEDDED_WASM_BASE64 = "${wasmBase64}";\n`
);

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

console.log('✨ Build finished successfully! Outputs:');
console.log('   - ESM: dist/index.mjs');
console.log('   - CJS: dist/index.cjs');
console.log('   - Types: dist/index.d.ts');
