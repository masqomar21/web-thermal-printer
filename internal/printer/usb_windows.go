//go:build windows

package printer

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procCreateFileW          = kernel32.NewProc("CreateFileW")
	procCloseHandle          = kernel32.NewProc("CloseHandle")
	winusb                   = syscall.NewLazyDLL("winusb.dll")
	procWinUsbInitialize     = winusb.NewProc("WinUsb_Initialize")
	procWinUsbFree           = winusb.NewProc("WinUsb_Free")
	procWinUsbQueryPipe      = winusb.NewProc("WinUsb_QueryPipe")
	procWinUsbWritePipe      = winusb.NewProc("WinUsb_WritePipe")
)

const (
	GENERIC_WRITE         = 0x40000000
	GENERIC_READ          = 0x80000000
	FILE_SHARE_READ       = 0x00000001
	FILE_SHARE_WRITE      = 0x00000002
	OPEN_EXISTING         = 3
	FILE_ATTRIBUTE_NORMAL = 0x80
	FILE_FLAG_OVERLAPPED  = 0x40000000
)

type USB_PIPE_INFORMATION struct {
	PipeType          uint32
	PipeId            byte
	MaximumPacketSize uint16
	Interval          byte
}

func writeUSBDataWindows(path string, data []byte) (int, error) {
	// If path is a USB raw device interface (\\?\USB#...), try WinUSB DLL write first
	if strings.HasPrefix(strings.ToUpper(path), `\\?\USB#`) || strings.HasPrefix(strings.ToUpper(path), `\??\USB#`) {
		n, err := writeWinUSB(path, data)
		if err == nil {
			return n, nil
		}
		log.Printf("⚠️ WinUSB write failed (%v), trying raw Win32 WriteFile...", err)
	}

	// Standard Win32 File Write
	return writeWin32File(path, data)
}

func writeWinUSB(devicePath string, data []byte) (int, error) {
	pathPtr, err := syscall.UTF16PtrFromString(devicePath)
	if err != nil {
		return 0, err
	}

	hDevice, _, lastErr := procCreateFileW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(GENERIC_READ|GENERIC_WRITE),
		uintptr(FILE_SHARE_READ|FILE_SHARE_WRITE),
		0,
		uintptr(OPEN_EXISTING),
		uintptr(FILE_ATTRIBUTE_NORMAL|FILE_FLAG_OVERLAPPED),
		0,
	)

	if hDevice == uintptr(syscall.InvalidHandle) || hDevice == 0 {
		return 0, fmt.Errorf("CreateFileW failed: %v", lastErr)
	}
	defer procCloseHandle.Call(hDevice)

	var hWinUSB uintptr
	ret, _, err := procWinUsbInitialize.Call(
		hDevice,
		uintptr(unsafe.Pointer(&hWinUSB)),
	)
	if ret == 0 {
		return 0, fmt.Errorf("WinUsb_Initialize failed: %v", err)
	}
	defer procWinUsbFree.Call(hWinUSB)

	if len(data) == 0 {
		return 0, nil
	}

	// Discover Bulk OUT Pipe ID
	var outPipeID byte = 0x01
	for i := byte(0); i < 10; i++ {
		var pipeInfo USB_PIPE_INFORMATION
		res, _, _ := procWinUsbQueryPipe.Call(
			hWinUSB,
			0,
			uintptr(i),
			uintptr(unsafe.Pointer(&pipeInfo)),
		)
		if res != 0 {
			if pipeInfo.PipeType == 2 && (pipeInfo.PipeId&0x80) == 0 {
				outPipeID = pipeInfo.PipeId
				break
			}
		}
	}

	var written uint32
	ret, _, err = procWinUsbWritePipe.Call(
		hWinUSB,
		uintptr(outPipeID),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(unsafe.Pointer(&written)),
		0,
	)

	if ret == 0 {
		return 0, fmt.Errorf("WinUsb_WritePipe failed: %v", err)
	}

	return int(written), nil
}

func writeWin32File(path string, data []byte) (int, error) {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		f, err = os.OpenFile(path, os.O_RDWR, 0)
	}
	if err != nil {
		return 0, err
	}
	defer f.Close()

	if len(data) == 0 {
		return 0, nil
	}

	return f.Write(data)
}
