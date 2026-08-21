@echo off
echo Building and Starting Antrean Ticket Printer Go Service...
go build -o printerRunner.exe ./cmd
if %ERRORLEVEL% EQU 0 (
    printerRunner.exe
) else (
    echo Build failed!
)
pause