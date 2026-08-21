package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type SocketConfig struct {
	URL                     string `json:"url"`
	TopicPrintNomorAntrean string `json:"topic_print_nomor_antrean"`
	TopicStatus             string `json:"topic_status"`
	ReconnectIntervalMs     int    `json:"reconnect_interval_ms"`
}

type PrinterConfig struct {
	Type           string `json:"type"`             // "usb", "net", "serial", "system"
	VendorID       uint16 `json:"vendor_id"`        // e.g. 0x04b8 (Epson) or 0x0fe6
	ProductID      uint16 `json:"product_id"`       // e.g. 0x0e15 or 0x811e
	IPAddress      string `json:"ip_address"`       // e.g. "192.168.1.200:9100"
	SerialPort     string `json:"serial_port"`      // e.g. "COM3" or "/dev/ttyUSB0"
	BaudRate       int    `json:"baud_rate"`        // e.g. 9600, 115200
	SystemPrinter  string `json:"system_printer_name"`
	RenderMode     string `json:"render_mode"`      // "text" or "image"
	PaperWidthDots int    `json:"paper_width_dots"` // 384 (58mm) or 576 (80mm)
}

type AppConfig struct {
	Socket  SocketConfig  `json:"socket"`
	Printer PrinterConfig `json:"printer"`
	WebURL  string        `json:"web_url"`
}

func LoadConfig() (*AppConfig, error) {
	// Try loading .env first if present
	_ = godotenv.Load()

	cfg := &AppConfig{
		WebURL: getEnv("WEB_URL", ""),
		Socket: SocketConfig{
			URL:                     getEnv("SOCKET_URL", "https://be.simpuskes.com"),
			TopicPrintNomorAntrean: getEnv("TOPIC_PRINT_NOMOR_ANTREAN", "antrean_print"),
			TopicStatus:             getEnv("TOPIC_STATUS", "status"),
			ReconnectIntervalMs:     getEnvInt("RECONNECT_INTERVAL_MS", 5000),
		},
		Printer: PrinterConfig{
			Type:           getEnv("PRINTER_TYPE", "usb"),
			VendorID:       getEnvHex16("PRINTER_VENDOR_ID", 0),
			ProductID:      getEnvHex16("PRINTER_PRODUCT_ID", 0),
			IPAddress:      getEnv("PRINTER_IP", "192.168.1.200:9100"),
			SerialPort:     getEnv("PRINTER_SERIAL_PORT", "COM3"),
			BaudRate:       getEnvInt("PRINTER_BAUD_RATE", 9600),
			SystemPrinter:  getEnv("PRINTER_SYSTEM_NAME", ""),
			RenderMode:     getEnv("PRINTER_RENDER_MODE", "both"), // "text", "image", or "both"
			PaperWidthDots: getEnvInt("PRINTER_PAPER_WIDTH", 576),   // Default 80mm printer (576 dots)
		},
	}

	// Try reading config.json to override defaults if config.json exists
	if _, err := os.Stat("config.json"); err == nil {
		data, err := os.ReadFile("config.json")
		if err == nil {
			var jsonCfg AppConfig
			if err := json.Unmarshal(data, &jsonCfg); err == nil {
				if jsonCfg.Socket.URL != "" {
					cfg.Socket.URL = jsonCfg.Socket.URL
				}
				if jsonCfg.Socket.TopicPrintNomorAntrean != "" {
					cfg.Socket.TopicPrintNomorAntrean = jsonCfg.Socket.TopicPrintNomorAntrean
				}
				if jsonCfg.Socket.TopicStatus != "" {
					cfg.Socket.TopicStatus = jsonCfg.Socket.TopicStatus
				}
				if jsonCfg.Printer.Type != "" {
					cfg.Printer.Type = jsonCfg.Printer.Type
				}
				if jsonCfg.Printer.RenderMode != "" {
					cfg.Printer.RenderMode = jsonCfg.Printer.RenderMode
				}
			}
		}
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvHex16(key string, defaultVal uint16) uint16 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	val = strings.TrimPrefix(val, "0x")
	val = strings.TrimPrefix(val, "0X")
	u, err := strconv.ParseUint(val, 16, 16)
	if err != nil {
		return defaultVal
	}
	return uint16(u)
}
