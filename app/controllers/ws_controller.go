package controllers

import (
	"encoding/json"
	"log"

	"backend-chacotuto/app/middleware"
	"backend-chacotuto/pkg/websocket"

	"github.com/gofiber/fiber/v2"
	fiberWebsocket "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// WsController encapsula nuestro Hub para no usar variables globales
type WsController struct {
	Hub *websocket.Hub
}

// NewWsController crea la instancia del controlador
func NewWsController(hub *websocket.Hub) *WsController {
	return &WsController{Hub: hub}
}

// UpgradeDroneMiddleware valida la solicitud de upgrade para drones (sin auth)
func (wsCtrl *WsController) UpgradeDroneMiddleware(c *fiber.Ctx) error {
	if fiberWebsocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// UpgradeGCSMiddleware valida la solicitud de upgrade para GCS (con auth JWT)
func (wsCtrl *WsController) UpgradeGCSMiddleware(c *fiber.Ctx) error {
	if !fiberWebsocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	// Extraer JWT del query parameter ?token=<JWT>
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token requerido en query parameter: ?token=<JWT>",
		})
	}

	// Validar JWT
	claims, err := middleware.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token inválido o expirado",
		})
	}

	// Inyectar datos del usuario
	c.Locals("userId", claims.UserID)
	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)
	c.Locals("allowed", true)

	return c.Next()
}

// ServeWSDrone es el handler WebSocket para conexiones de drones
func (wsCtrl *WsController) ServeWSDrone(conn *fiberWebsocket.Conn) {
	clientID := uuid.New().String()

	// Crear cliente tipo drone
	client := websocket.NewClient(clientID, websocket.ClientTypeDrone, conn, wsCtrl.Hub)

	log.Printf("🛩️ Nueva conexión de dron (WS ID: %s)", clientID)

	// Registrar en el hub
	wsCtrl.Hub.RegisterDroneClient(client)

	// Ejecutar lectura y escritura
	go client.Write()
	client.ReadDrone() // Bloquea hasta que se desconecte
}

// ServeWSGCS es el handler WebSocket para conexiones del GCS (dashboard)
func (wsCtrl *WsController) ServeWSGCS(conn *fiberWebsocket.Conn) {
	// Usar el username + UUID único para permitir múltiples pestañas por usuario
	username := conn.Locals("username")
	clientID := "gcs-"
	if u, ok := username.(string); ok && u != "" {
		clientID += u + "-" + uuid.New().String()
	} else {
		clientID += uuid.New().String()
	}

	// Crear cliente tipo GCS
	client := websocket.NewClient(clientID, websocket.ClientTypeGCS, conn, wsCtrl.Hub)

	log.Printf("🎮 Nueva conexión GCS: %s", clientID)

	// Registrar en el hub
	wsCtrl.Hub.RegisterGCSClient(client)

	// Enviar snapshot inicial al GCS con el estado actual de todos los drones
	registry := wsCtrl.Hub.GetRegistry()
	drones := registry.GetAllDrones()

	initialState := map[string]interface{}{
		"type":   "initial_state",
		"drones": drones,
	}
	if data, err := json.Marshal(initialState); err == nil {
		select {
		case client.GetSendChannel() <- data:
		default:
		}
	}

	// Ejecutar lectura y escritura
	go client.Write()
	client.ReadGCS() // Bloquea hasta que se desconecte
}
