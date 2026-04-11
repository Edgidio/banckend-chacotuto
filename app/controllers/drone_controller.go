package controllers

import (
	"backend-chacotuto/app/models"
	"backend-chacotuto/pkg/database"
	"backend-chacotuto/pkg/websocket"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// DroneController maneja las rutas REST de drones
type DroneController struct {
	Hub *websocket.Hub
}

// NewDroneController crea la instancia del controlador
func NewDroneController(hub *websocket.Hub) *DroneController {
	return &DroneController{Hub: hub}
}

// GetAll lista todos los drones registrados con su estado actual
func (dc *DroneController) GetAll(c *fiber.Ctx) error {
	// Obtener drones de la BD
	var drones []models.Drone
	database.DB.Order("created_at DESC").Find(&drones)

	// Enriquecer con estado en memoria (online/offline en tiempo real)
	registry := dc.Hub.GetRegistry()
	result := make([]fiber.Map, 0, len(drones))

	for _, drone := range drones {
		d := fiber.Map{
			"droneId":       drone.DroneID,
			"name":          drone.Name,
			"status":        drone.Status,
			"lastHeartbeat": drone.LastHeartbeat,
			"lastLat":       drone.LastLat,
			"lastLng":       drone.LastLng,
			"lastAlt":       drone.LastAlt,
			"isOnline":      drone.IsOnline,
			"createdAt":     drone.CreatedAt,
		}

		// Si está online, sobreescribir con datos en tiempo real
		if liveData, ok := registry.GetDrone(drone.DroneID); ok {
			d["status"] = liveData["status"]
			d["isOnline"] = liveData["isOnline"]
			d["lastHeartbeat"] = liveData["lastHeartbeat"]
			if telem, ok := liveData["lastTelemetry"]; ok {
				d["lastTelemetry"] = telem
			}
			if mission, ok := liveData["currentMission"]; ok {
				d["currentMission"] = mission
			}
		}

		result = append(result, d)
	}

	return c.JSON(fiber.Map{
		"drones": result,
		"total":  len(result),
	})
}

// GetOne obtiene un dron específico por su ID
func (dc *DroneController) GetOne(c *fiber.Ctx) error {
	droneID := c.Params("id")

	var drone models.Drone
	result := database.DB.Where("drone_id = ?", droneID).First(&drone)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Dron no encontrado",
		})
	}

	d := fiber.Map{
		"droneId":       drone.DroneID,
		"name":          drone.Name,
		"status":        drone.Status,
		"lastHeartbeat": drone.LastHeartbeat,
		"lastLat":       drone.LastLat,
		"lastLng":       drone.LastLng,
		"lastAlt":       drone.LastAlt,
		"isOnline":      drone.IsOnline,
		"createdAt":     drone.CreatedAt,
	}

	// Enriquecer con datos en tiempo real
	registry := dc.Hub.GetRegistry()
	if liveData, ok := registry.GetDrone(droneID); ok {
		d["status"] = liveData["status"]
		d["isOnline"] = liveData["isOnline"]
		d["lastHeartbeat"] = liveData["lastHeartbeat"]
		if telem, ok := liveData["lastTelemetry"]; ok {
			d["lastTelemetry"] = telem
		}
	}

	// Obtener misiones del dron
	var missions []models.Mission
	database.DB.Where("drone_id = ?", droneID).Order("created_at DESC").Find(&missions)
	d["missions"] = missions

	return c.JSON(d)
}

// GetTelemetry obtiene el historial de telemetría de un dron
func (dc *DroneController) GetTelemetry(c *fiber.Ctx) error {
	droneID := c.Params("id")

	// Paginación
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	missionID := c.Query("mission", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 100
	}

	offset := (page - 1) * limit

	query := database.DB.Where("drone_id = ?", droneID)
	if missionID != "" {
		query = query.Where("mission_id = ?", missionID)
	}

	var total int64
	query.Model(&models.TelemetryLog{}).Count(&total)

	var logs []models.TelemetryLog
	query.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&logs)

	return c.JSON(fiber.Map{
		"droneId":  droneID,
		"page":     page,
		"limit":    limit,
		"total":    total,
		"telemetry": logs,
	})
}
