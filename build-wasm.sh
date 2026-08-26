#!/bin/bash
set -e

echo "🔨 Building Thermal Printer Go WASM..."
mkdir -p web

# Copy runtime glue if missing
if [ ! -f "web/wasm_exec.js" ]; then
    GOROOT_PATH=$(go env GOROOT)
    if [ -f "$GOROOT_PATH/lib/wasm/wasm_exec.js" ]; then
        cp "$GOROOT_PATH/lib/wasm/wasm_exec.js" web/wasm_exec.js
    elif [ -f "$GOROOT_PATH/misc/wasm/wasm_exec.js" ]; then
        cp "$GOROOT_PATH/misc/wasm/wasm_exec.js" web/wasm_exec.js
    fi
fi

mkdir -p dist
GOOS=js GOARCH=wasm go build -v -o web/printer.wasm ./cmd/wasm-printer
cp web/printer.wasm dist/printer.wasm

echo "✅ Build successful! Output: web/printer.wasm and dist/printer.wasm"

