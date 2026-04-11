package controllers

import (
	"fmt"
	"time"

	"backend-chacotuto/app/models"
	"backend-chacotuto/pkg/database"
	"backend-chacotuto/pkg/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MissionController maneja las rutas REST de misiones
type MissionController struct {
	Hub *websocket.Hub
}

// NewMissionController crea la instancia del controlador
func NewMissionController(hub *websocket.Hub) *MissionController {
	return &MissionController{Hub: hub}
}

// CreateMissionRequest es el cuerpo para crear una misión
type CreateMissionRequest struct {
	DroneID    string             `json:"droneId"`
	MissionID  string             `json:"missionId,omitempty"` // Opcional, se genera si vacío
	Waypoints  []models.WaypointMsg `json:"waypoints"`
	StartPoint *models.WaypointMsg  `json:"startPoint,omitempty"`
}

// GetAll lista todas las misiones
func (mc *MissionController) GetAll(c *fiber.Ctx) error {
	droneID := c.Query("droneId")

	var missions []models.Mission
	query := database.DB.Preload("Waypoints").Order("created_at DESC")
	
	if droneID != "" {
		query = query.Where("drone_id = ?", droneID)
	}
	
	query.Find(&missions)

	return c.JSON(fiber.Map{
		"missions": missions,
		"total":    len(missions),
	})
}

// GetOne obtiene una misión específica
func (mc *MissionController) GetOne(c *fiber.Ctx) error {
	missionID := c.Params("id")

	var mission models.Mission
	result := database.DB.Preload("Waypoints").Where("mission_id = ?", missionID).First(&mission)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Misión no encontrada",
		})
	}

	// Obtener telemetría asociada
	var telemetryCount int64
	database.DB.Model(&models.TelemetryLog{}).Where("mission_id = ?", missionID).Count(&telemetryCount)

	return c.JSON(fiber.Map{
		"mission":        mission,
		"telemetryCount": telemetryCount,
	})
}

// Create crea y asigna una misión a un dron, enviándola por WebSocket
func (mc *MissionController) Create(c *fiber.Ctx) error {
	var req CreateMissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Datos de misión inválidos",
		})
	}

	if req.DroneID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "droneId es requerido",
		})
	}

	if len(req.Waypoints) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Se requiere al menos un waypoint",
		})
	}

	// Generar missionId si no se proporcionó
	if req.MissionID == "" {
		req.MissionID = fmt.Sprintf("MISSION-%s", uuid.New().String()[:8])
	}

	// Verificar que el dron existe y está online
	registry := mc.Hub.GetRegistry()
	if !registry.IsDroneOnline(req.DroneID) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El dron no está conectado",
		})
	}

	// Crear misión en BD
	mission := models.Mission{
		MissionID: req.MissionID,
		DroneID:   req.DroneID,
		Status:    "pending",
	}

	if req.StartPoint != nil {
		mission.StartLat = req.StartPoint.Lat
		mission.StartLng = req.StartPoint.Lng
		mission.StartAlt = req.StartPoint.Alt
	} else if len(req.Waypoints) > 0 {
		mission.StartLat = req.Waypoints[0].Lat
		mission.StartLng = req.Waypoints[0].Lng
		mission.StartAlt = req.Waypoints[0].Alt
	}

	if result := database.DB.Create(&mission); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creando misión en BD",
		})
	}

	// Guardar waypoints en BD
	for _, wp := range req.Waypoints {
		waypoint := models.Waypoint{
			MissionID: req.MissionID,
			Lat:       wp.Lat,
			Lng:       wp.Lng,
			Alt:       wp.Alt,
			Action:    wp.Action,
			Index:     wp.Index,
		}
		database.DB.Create(&waypoint)
	}

	// Construir mensaje mission_assign según el protocolo
	assignMsg := models.MissionAssignMsg{
		Type:      "mission_assign",
		MissionID: req.MissionID,
		Waypoints: req.Waypoints,
		StartPoint: req.StartPoint,
	}

	// Enviar por WebSocket al dron
	sent := mc.Hub.SendToDrone(req.DroneID, assignMsg)
	if !sent {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "No se pudo enviar la misión al dron",
			"mission": mission,
		})
	}

	// Notificar al GCS
	mc.Hub.BroadcastToGCS(map[string]interface{}{
		"type":      "mission_assigned",
		"missionId": req.MissionID,
		"droneId":   req.DroneID,
		"waypoints": req.Waypoints,
		"timestamp": time.Now().UnixMilli(),
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Misión creada y enviada al dron",
		"mission": mission,
	})
}

// Cancel cancela una misión activa
func (mc *MissionController) Cancel(c *fiber.Ctx) error {
	missionID := c.Params("id")

	var mission models.Mission
	result := database.DB.Where("mission_id = ?", missionID).First(&mission)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Misión no encontrada",
		})
	}

	if mission.Status == "completed" || mission.Status == "cancelled" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "La misión ya está " + mission.Status,
		})
	}

	// Actualizar en BD
	database.DB.Model(&mission).Update("status", "cancelled")

	// Enviar mission_cancel al dron
	cancelMsg := models.MissionCancelMsg{
		Type:      "mission_cancel",
		MissionID: missionID,
		Reason:    "Cancelada por el operador",
	}
	mc.Hub.SendToDrone(mission.DroneID, cancelMsg)

	// Completar misión en registro
	registry := mc.Hub.GetRegistry()
	registry.CompleteMission(mission.DroneID, missionID)

	// Notificar al GCS
	mc.Hub.BroadcastToGCS(map[string]interface{}{
		"type":      "mission_cancelled",
		"missionId": missionID,
		"droneId":   mission.DroneID,
	})

	return c.JSON(fiber.Map{
		"message": "Misión cancelada",
	})
}
