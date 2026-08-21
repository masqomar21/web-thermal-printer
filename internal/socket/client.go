package socket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/masqomar21/antrean-ticket-printer/internal/config"
	"github.com/masqomar21/antrean-ticket-printer/internal/model"
)

type TicketHandlerFunc func(data model.TicketData) error

type Client struct {
	cfg          config.SocketConfig
	conn         *websocket.Conn
	mu           sync.Mutex
	isConnected  bool
	handler      TicketHandlerFunc
	stopChan     chan struct{}
	reconnectSec time.Duration
}

func NewClient(cfg config.SocketConfig, handler TicketHandlerFunc) *Client {
	interval := cfg.ReconnectIntervalMs
	if interval <= 0 {
		interval = 5000
	}
	return &Client{
		cfg:          cfg,
		handler:      handler,
		stopChan:     make(chan struct{}),
		reconnectSec: time.Duration(interval) * time.Millisecond,
	}
}

func (c *Client) Start() {
	go c.connectLoop()
}

func (c *Client) Stop() {
	close(c.stopChan)
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.mu.Unlock()
}

func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected
}

func (c *Client) connectLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
			log.Printf("🔌 Connecting to Socket.IO server at %s ...", c.cfg.URL)
			if err := c.connect(); err != nil {
				log.Printf("❌ Socket connection failed: %v. Retrying in %v...", err, c.reconnectSec)
				time.Sleep(c.reconnectSec)
				continue
			}
			log.Println("✅ Connected to Socket.IO server!")

			// Read loop
			c.readLoop()

			c.mu.Lock()
			c.isConnected = false
			c.mu.Unlock()

			log.Printf("🔴 Connection lost. Reconnecting in %v...", c.reconnectSec)
			time.Sleep(c.reconnectSec)
		}
	}
}

func (c *Client) connect() error {
	u, err := url.Parse(c.cfg.URL)
	if err != nil {
		return fmt.Errorf("invalid socket URL: %w", err)
	}

	scheme := "ws"
	if u.Scheme == "https" {
		scheme = "wss"
	}
	u.Scheme = scheme
	if !strings.HasSuffix(u.Path, "/socket.io/") {
		u.Path = strings.TrimSuffix(u.Path, "/") + "/socket.io/"
	}

	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.isConnected = true
	c.mu.Unlock()

	// Send Engine.IO / Socket.IO initial connect message
	if err := conn.WriteMessage(websocket.TextMessage, []byte("40")); err != nil {
		conn.Close()
		return err
	}

	// Send status connected event
	c.EmitStatus("connected", "Go Ticket Printer Connected")

	return nil
}

func (c *Client) readLoop() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("⚠️ Socket read error: %v", err)
			return
		}

		msgStr := string(message)
		c.handlePacket(msgStr)
	}
}

func (c *Client) handlePacket(packet string) {
	if packet == "2" { // Engine.IO Ping
		c.sendRaw("3") // Engine.IO Pong
		return
	}

	// Socket.IO event format: 42[id]["event_name", payload] or 42["event_name", payload]
	if strings.HasPrefix(packet, "42") {
		c.parseAndHandleEvent(packet[2:])
	}
}

var ackRegex = regexp.MustCompile(`^(\d+)\[(.+)\]$`)

func (c *Client) parseAndHandleEvent(payload string) {
	var ackID string
	dataStr := payload

	// Check if there is an ACK ID attached e.g. 1["antrean_print", {...}]
	if match := ackRegex.FindStringSubmatch(payload); len(match) == 3 {
		ackID = match[1]
		dataStr = "[" + match[2] + "]"
	}

	var rawArray []json.RawMessage
	if err := json.Unmarshal([]byte(dataStr), &rawArray); err != nil || len(rawArray) < 1 {
		return
	}

	var eventName string
	if err := json.Unmarshal(rawArray[0], &eventName); err != nil {
		return
	}

	log.Printf("📩 Socket event received: %s", eventName)

	if eventName == c.cfg.TopicPrintNomorAntrean || eventName == "antrean_print" {
		if len(rawArray) > 1 {
			var ticket model.TicketData
			if err := json.Unmarshal(rawArray[1], &ticket); err == nil {
				log.Printf("🖨️ Ticket print payload: %+v", ticket)

				// Invoke print handler
				if c.handler != nil {
					err := c.handler(ticket)
					if err != nil {
						log.Printf("❌ Print error: %v", err)
						c.EmitStatus("error", err.Error())
					} else {
						log.Println("✅ Print completed successfully")
						c.EmitStatus("printed", "")
					}
				}

				// Send ACK response back if requested
				if ackID != "" {
					c.sendAck(ackID, "ok")
				}
			} else {
				log.Printf("⚠️ Failed to parse ticket payload: %v", err)
			}
		}
	}
}

func (c *Client) EmitStatus(status string, message string) {
	payload := model.PrintStatusPayload{
		Status:  status,
		Message: message,
	}

	data, err := json.Marshal([]interface{}{c.cfg.TopicStatus, payload})
	if err != nil {
		return
	}

	packet := "42" + string(data)
	c.sendRaw(packet)
}

func (c *Client) sendAck(ackID string, response string) {
	data, _ := json.Marshal([]interface{}{response})
	packet := "43" + ackID + string(data)
	c.sendRaw(packet)
	log.Printf("📤 Sent ACK (%s) to server", ackID)
}

func (c *Client) sendRaw(msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && c.isConnected {
		_ = c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}
