package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hoshea/orion-backend/internal/domain"
)

// Connection WebSocket 连接封装
type Connection struct {
	ID         string
	conn       *websocket.Conn
	send       chan []byte
	mu         sync.Mutex
	closed     bool
	pingTicker *time.Ticker
}

// NewConnection 创建新连接
func NewConnection(id string, conn *websocket.Conn) *Connection {
	return &Connection{
		ID:         id,
		conn:       conn,
		send:       make(chan []byte, 256),
		pingTicker: time.NewTicker(30 * time.Second),
	}
}

// ReadPump 读取客户端消息
func (c *Connection) ReadPump(handleMessage func([]byte)) {
	defer func() {
		c.Close()
	}()

	// 设置读取超时
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		handleMessage(message)
	}
}

// WritePump 发送消息给客户端
func (c *Connection) WritePump() {
	defer func() {
		c.pingTicker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel 关闭
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-c.pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage 发送消息
func (c *Connection) SendMessage(msg *domain.WebSocketMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	msg.Timestamp = time.Now()
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		// Channel 已满
		log.Printf("Warning: send buffer full for connection %s", c.ID)
		return nil
	}
}

// SendJSON 发送 JSON 消息（简化版）
func (c *Connection) SendJSON(messageType domain.MessageType, payload interface{}) error {
	return c.SendMessage(&domain.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
	})
}

// Close 关闭连接
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.send)
		c.conn.Close()
		log.Printf("Connection %s closed", c.ID)
	}
}

// IsClosed 检查连接是否已关闭
func (c *Connection) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}
