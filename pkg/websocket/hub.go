package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maneja dos tipos de clientes: drones y GCS
type Hub struct {
	// Clientes GCS (dashboard web)
	gcsClients map[string]*Client
	gcsMu      sync.RWMutex

	// Canales de control
	registerDrone   chan *Client
	unregisterDrone chan *Client
	registerGCS     chan *Client
	unregisterGCS   chan *Client

	// Protocolo
	droneMessage chan *RawDroneMessage
	registry     *DroneRegistry
	handler      *ProtocolHandler
}

// RawDroneMessage empaqueta un mensaje crudo del dron con su cliente
type RawDroneMessage struct {
	Client *Client
	Data   []byte
}

func NewHub() *Hub {
	h := &Hub{
		gcsClients:      make(map[string]*Client),
		registerDrone:   make(chan *Client),
		unregisterDrone: make(chan *Client),
		registerGCS:     make(chan *Client),
		unregisterGCS:   make(chan *Client),
		droneMessage:    make(chan *RawDroneMessage, 256),
	}

	h.registry = NewDroneRegistry(h)
	h.handler = NewProtocolHandler(h, h.registry)

	return h
}

// GetRegistry expone el registro de drones
func (h *Hub) GetRegistry() *DroneRegistry {
	return h.registry
}

// RegisterDroneClient encola un alta de drone
func (h *Hub) RegisterDroneClient(c *Client) {
	h.registerDrone <- c
}

// UnregisterDroneClient encola una baja de drone
func (h *Hub) UnregisterDroneClient(c *Client) {
	h.unregisterDrone <- c
}

// RegisterGCSClient encola un alta de GCS
func (h *Hub) RegisterGCSClient(c *Client) {
	h.registerGCS <- c
}

// UnregisterGCSClient encola una baja de GCS
func (h *Hub) UnregisterGCSClient(c *Client) {
	h.unregisterGCS <- c
}

// HandleDroneMessage encola un mensaje de un dron para procesarlo
func (h *Hub) HandleDroneMessage(client *Client, data []byte) {
	h.droneMessage <- &RawDroneMessage{Client: client, Data: data}
}

// BroadcastToGCS envía un mensaje a TODOS los clientes GCS conectados
func (h *Hub) BroadcastToGCS(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error serializando mensaje para GCS:", err)
		return
	}

	h.gcsMu.RLock()
	defer h.gcsMu.RUnlock()

	for id, client := range h.gcsClients {
		select {
		case client.send <- data:
		default:
			log.Printf("GCS client %s con buffer lleno, desconectando", id)
			close(client.send)
			delete(h.gcsClients, id)
		}
	}
}

// SendToDrone envía un mensaje a un dron específico por su droneId
func (h *Hub) SendToDrone(droneID string, msg interface{}) bool {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error serializando mensaje para dron:", err)
		return false
	}

	h.registry.mu.RLock()
	drone, exists := h.registry.drones[droneID]
	h.registry.mu.RUnlock()

	if !exists || drone.Client == nil {
		return false
	}

	select {
	case drone.Client.send <- data:
		return true
	default:
		log.Printf("Dron %s con buffer lleno", droneID)
		return false
	}
}

// Run es el loop principal del Hub — procesa todos los eventos
func (h *Hub) Run() {
	// Iniciar el monitor de heartbeats
	go h.registry.StartHeartbeatMonitor()

	for {
		select {
		// --- Drones ---
		case client := <-h.registerDrone:
			log.Printf("🛩️ Dron conectado (WS): %s", client.ID)
			// El registro real ocurre cuando el dron envía "register"

		case client := <-h.unregisterDrone:
			log.Printf("🛩️ Dron desconectado (WS): %s", client.ID)
			droneID := client.DroneID
			if droneID != "" {
				h.registry.MarkOffline(droneID)
				h.BroadcastToGCS(map[string]interface{}{
					"type":    "drone_offline",
					"droneId": droneID,
				})
			}
			close(client.send)

		// --- GCS ---
		case client := <-h.registerGCS:
			h.gcsMu.Lock()
			h.gcsClients[client.ID] = client
			h.gcsMu.Unlock()
			log.Printf("🎮 GCS conectado: %s", client.ID)

		case client := <-h.unregisterGCS:
			h.gcsMu.Lock()
			if _, ok := h.gcsClients[client.ID]; ok {
				delete(h.gcsClients, client.ID)
				close(client.send)
			}
			h.gcsMu.Unlock()
			log.Printf("🎮 GCS desconectado: %s", client.ID)

		// --- Mensajes de drones ---
		case rawMsg := <-h.droneMessage:
			h.handler.HandleDroneMessage(rawMsg.Client, rawMsg.Data)
		}
	}
}
