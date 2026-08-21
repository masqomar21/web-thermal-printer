//go:build !windows

package printer

import (
	"os"
)

func writeUSBDataWindows(path string, data []byte) (int, error) {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		f, err = os.OpenFile(path, os.O_RDWR, 0)
	}
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Write(data)
}
