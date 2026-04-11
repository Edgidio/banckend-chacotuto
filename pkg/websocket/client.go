package websocket

import (
	"log"

	fiberWebsocket "github.com/gofiber/websocket/v2"
)

// ClientType distingue entre conexiones de dron y GCS
const (
	ClientTypeDrone = "drone"
	ClientTypeGCS   = "gcs"
)

// Client representa una conexión WebSocket activa
type Client struct {
	ID         string // UUID o droneId
	DroneID    string // Solo para clientes tipo drone (se llena después del "register")
	ClientType string // "drone" o "gcs"
	hub        *Hub
	conn       *fiberWebsocket.Conn
	send       chan []byte
}

// NewClient crea un nuevo cliente WebSocket
func NewClient(id string, clientType string, conn *fiberWebsocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:         id,
		ClientType: clientType,
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
	}
}

// GetSendChannel expone el canal de envío para uso directo
func (c *Client) GetSendChannel() chan []byte {
	return c.send
}

// Write envía mensajes del canal al WebSocket del cliente
func (c *Client) Write() {
	defer func() {
		_ = c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok {
			_ = c.conn.WriteMessage(fiberWebsocket.CloseMessage, []byte{})
			return
		}
		err := c.conn.WriteMessage(fiberWebsocket.TextMessage, message)
		if err != nil {
			log.Printf("Error enviando mensaje al cliente %s: %v", c.ID, err)
			break
		}
	}
}

// ReadDrone lee mensajes de un dron y los envía al hub para procesamiento
func (c *Client) ReadDrone() {
	defer func() {
		c.hub.UnregisterDroneClient(c)
		_ = c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if fiberWebsocket.IsUnexpectedCloseError(err, fiberWebsocket.CloseGoingAway, fiberWebsocket.CloseAbnormalClosure) {
				log.Printf("Dron desconectado (%s) por error: %v", c.ID, err)
			}
			break
		}
		c.hub.HandleDroneMessage(c, message)
	}
}

// ReadGCS lee mensajes del GCS (por ahora solo log, el GCS usa REST para enviar comandos)
func (c *Client) ReadGCS() {
	defer func() {
		c.hub.UnregisterGCSClient(c)
		_ = c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if fiberWebsocket.IsUnexpectedCloseError(err, fiberWebsocket.CloseGoingAway, fiberWebsocket.CloseAbnormalClosure) {
				log.Printf("GCS desconectado (%s) por error: %v", c.ID, err)
			}
			break
		}
		// El GCS no envía mensajes por WebSocket, solo recibe broadcasts
		// Los comandos van por REST API
	}
}
