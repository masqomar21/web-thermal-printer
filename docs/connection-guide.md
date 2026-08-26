# Connection Guide: Web USB vs Web Serial

This guide explains how browser hardware connections work in `thermal-printer-wasm` and how to troubleshoot connection dialogs and security policies.

---

## 1. Web USB vs Web Serial: Differences

| Feature | Web USB (`navigator.usb`) | Web Serial (`navigator.serial`) |
|---|---|---|
| **Target Hardware** | Direct USB receipt printers | Virtual COM / Serial ports (`/dev/cu.usbserial`, `COM3`) |
| **Interface Class** | USB Class `0x07` (Printer) or Vendor Specific | USB CDC-ACM / Serial Chips (CH340, FTDI, PL2303) |
| **Browser Dialog Prompt** | Lists `USB Receipt Printer` | Lists Virtual Serial COM ports |
| **Recommendation** | **Recommended** for standard USB thermal printers | For RS232 Serial cable printers or Serial-to-USB converters |

---

## 2. Why Printer Appears in Web USB but NOT in Web Serial?

If your printer is connected with a standard USB cable:
- **Web USB** accesses the raw USB interface directly (`USB Class 0x07`). It reads the USB descriptor and displays **`USB Receipt Printer`**.
- **Web Serial** only lists devices registered as **Virtual COM Serial TTY ports** by macOS or Windows kernel (`/dev/cu.usbmodem...` or `/dev/cu.usbserial...`).
- Pure USB thermal receipt printers do not create Serial COM device nodes in macOS/Windows unless a special Virtual COM Mode or USB-to-Serial converter is used.

> 💡 **Rule of Thumb**: For standard USB thermal printers, **always use `.connectUSB(1)`**.

---

## 3. Browser Security & HTTPS Policies

Both Web USB and Web Serial APIs are classified as **Security-Sensitive Web APIs** by W3C and Chromium.

### Security Constraints:
1. **User Gesture Requirement**: `.connectUSB()` and `.connectSerial()` **MUST** be invoked as a direct result of a user gesture (e.g. inside a button `click` event listener). They cannot be called on page load.
2. **Secure Context (`HTTPS` or `localhost`)**:
   - Browsers disable `navigator.usb` and `navigator.serial` over unencrypted HTTP connections (`http://192.168.x.x`).

---

## 4. Local Network Testing Workaround (Chrome Flags)

If you are developing or testing over a local LAN IP address (e.g. `http://192.168.1.100:8080`):

1. Open Chrome on the client device.
2. Navigate to:
   ```text
   chrome://flags/#unsafely-treat-insecure-origin-as-secure
   ```
3. Set the flag to **Enabled**.
4. In the text area, enter your local origin URL (e.g. `http://192.168.1.100:8080`).
5. Click **Relaunch** in Chrome.

After relaunching, Web USB and Web Serial permission dialogs will function normally over local HTTP.
