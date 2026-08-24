@echo off
echo 🔨 Building Thermal Printer Go WASM for Windows...
if not exist "web" mkdir web

set GOOS=js
set GOARCH=wasm
go build -v -o web/printer.wasm ./cmd/wasm-printer

echo ✅ Build successful! Output: web/printer.wasm
