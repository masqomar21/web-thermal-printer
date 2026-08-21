@echo off
setlocal EnableDelayedExpansion

:: Detect if running in PowerShell
for /f "delims=" %%A in ('powershell -Command "$PSVersionTable.PSVersion.Major"') do set PS_VERSION=%%A

if defined PS_VERSION (
    echo Running in PowerShell...
    set USE_POWERSHELL=1
) else (
    echo Running in Command Prompt...
    set USE_POWERSHELL=0
)

echo Checking if Node.js is installed...
where node >nul 2>nul
IF %ERRORLEVEL% NEQ 0 (
    echo Node.js is not installed. Downloading and installing Node.js 22 LTS...
    
    if %USE_POWERSHELL%==1 (
        powershell -Command "Invoke-WebRequest -Uri 'https://nodejs.org/dist/v22.13.1/node-v22.13.1-x64.msi' -OutFile 'nodejs.msi'"
    ) else (
        curl -o nodejs.msi https://nodejs.org/dist/v22.13.1/node-v22.13.1-x64.msi
    )
    
    start /wait msiexec /i nodejs.msi
    del nodejs.msi
    echo Node.js installation complete.

    echo Please restart your terminal and run 'install_deps.bat' to install dependencies.
    pause
    exit
) ELSE (
    echo Node.js is already installed.
)

echo Node.js setup is complete.
pause
