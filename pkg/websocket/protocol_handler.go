package websocket

import (
	"encoding/json"
	"log"
	"time"

	"backend-chacotuto/app/models"
	"backend-chacotuto/pkg/database"
)

// ProtocolHandler implementa el switch del protocolo CHACOTUTO
type ProtocolHandler struct {
	hub      *Hub
	registry *DroneRegistry
}

// NewProtocolHandler crea el handler del protocolo
func NewProtocolHandler(hub *Hub, registry *DroneRegistry) *ProtocolHandler {
	return &ProtocolHandler{
		hub:      hub,
		registry: registry,
	}
}

// HandleDroneMessage procesa un mensaje crudo de un dron
func (p *ProtocolHandler) HandleDroneMessage(client *Client, raw []byte) {
	// Parsear el tipo de mensaje
	var base models.BaseMessage
	if err := json.Unmarshal(raw, &base); err != nil {
		log.Printf("Error parseando mensaje de dron: %v", err)
		return
	}

	switch base.Type {
	case "register":
		p.handleRegister(client, raw)
	case "heartbeat":
		p.handleHeartbeat(client, raw)
	case "telemetry":
		p.handleTelemetry(client, raw)
	case "mission_ack":
		p.handleMissionAck(client, raw)
	case "mission_ready":
		p.handleMissionReady(client, raw)
	case "mission_complete":
		p.handleMissionComplete(client, raw)
	default:
		log.Printf("Tipo de mensaje desconocido del dron: %s", base.Type)
	}
}

// handleRegister — Registro del dron al conectarse
func (p *ProtocolHandler) handleRegister(client *Client, raw []byte) {
	var msg models.RegisterMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("Error parseando register: %v", err)
		return
	}

	// Registrar en memoria
	p.registry.Register(msg.DroneID, client)

	// Persistir en BD (crear o actualizar)
	if database.DB != nil {
		drone := models.Drone{
			DroneID:       msg.DroneID,
			Name:          msg.DroneID, // Nombre por defecto = su ID
			Status:        "idle",
			IsOnline:      true,
			LastHeartbeat: time.Now(),
		}

		// Upsert: crear si no existe, actualizar si ya existe
		result := database.DB.Where("drone_id = ?", msg.DroneID).First(&models.Drone{})
		if result.RowsAffected == 0 {
			database.DB.Create(&drone)
			log.Printf("💾 Dron nuevo guardado en BD: %s", msg.DroneID)
		} else {
			database.DB.Model(&models.Drone{}).Where("drone_id = ?", msg.DroneID).Updates(map[string]interface{}{
				"status":         "idle",
				"is_online":      true,
				"last_heartbeat": time.Now(),
			})
		}
	}

	// Notificar al GCS
	p.hub.BroadcastToGCS(map[string]interface{}{
		"type":    "drone_online",
		"droneId": msg.DroneID,
	})

	log.Printf("✅ register procesado — droneId: %s", msg.DroneID)
}

// handleHeartbeat — Latido periódico cada 2 segundos
func (p *ProtocolHandler) handleHeartbeat(client *Client, raw []byte) {
	var msg models.HeartbeatMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("Error parseando heartbeat: %v", err)
		return
	}

	// Actualizar en memoria
	p.registry.UpdateHeartbeat(msg.DroneID, msg.Status)

	// Actualizar en BD (no en cada heartbeat, solo cada 10s)
	if database.DB != nil {
		database.DB.Model(&models.Drone{}).Where("drone_id = ?", msg.DroneID).Updates(map[string]interface{}{
			"status":         msg.Status,
			"is_online":      true,
			"last_heartbeat": time.Now(),
		})
	}

	// Reenviar al GCS
	p.hub.BroadcastToGCS(map[string]interface{}{
		"type":    "drone_status",
		"droneId": msg.DroneID,
		"status":  msg.Status,
	})
}

// handleTelemetry — Datos de sensores en tiempo real (cada 100ms durante misión)
func (p *ProtocolHandler) handleTelemetry(client *Client, raw []byte) {
	var msg models.TelemetryMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("Error parseando telemetry: %v", err)
		return
	}

	// Actualizar estado en memoria (cada mensaje)
	var rawMap map[string]interface{}
	json.Unmarshal(raw, &rawMap)
	p.registry.UpdateTelemetry(msg.DroneID, rawMap)

	// Actualizar última posición GPS en BD
	if database.DB != nil && msg.GPS != nil {
		database.DB.Model(&models.Drone{}).Where("drone_id = ?", msg.DroneID).Updates(map[string]interface{}{
			"last_lat": msg.GPS.Lat,
			"last_lng": msg.GPS.Lng,
			"last_alt": msg.GPS.Alt,
		})
	}

	// Obtener estado real del dron para saber si guardar o no
	droneInfo, exists := p.registry.GetDrone(msg.DroneID)
	isExecutingMission := exists && droneInfo["status"] == "in_mission"

	// Guardar en BD solo si está ejecutando la misión formalmente (no pre-vuelo)
	if database.DB != nil && p.registry.ShouldSaveTelemetry(msg.DroneID) && isExecutingMission {
		telemetryLog := models.TelemetryLog{
			DroneID:   msg.DroneID,
			Timestamp: msg.Timestamp,
			// Sensores
			GyroX:  msg.Sensors.Gyroscope.X,
			GyroY:  msg.Sensors.Gyroscope.Y,
			GyroZ:  msg.Sensors.Gyroscope.Z,
			AccelX: msg.Sensors.Accelerometer.X,
			AccelY: msg.Sensors.Accelerometer.Y,
			AccelZ: msg.Sensors.Accelerometer.Z,
			MagX:   msg.Sensors.Magnetometer.X,
			MagY:   msg.Sensors.Magnetometer.Y,
			MagZ:   msg.Sensors.Magnetometer.Z,
			// Orientación
			Pitch: msg.Orientation.Pitch,
			Roll:  msg.Orientation.Roll,
			Yaw:   msg.Orientation.Yaw,
		}

		// GPS puede ser null
		if msg.GPS != nil {
			telemetryLog.Lat = msg.GPS.Lat
			telemetryLog.Lng = msg.GPS.Lng
			telemetryLog.Alt = msg.GPS.Alt
		}

		// Progreso de misión
		if msg.Mission != nil {
			telemetryLog.CurrentWaypoint = msg.Mission.CurrentWaypointIndex
			telemetryLog.TotalWaypoints = msg.Mission.TotalWaypoints
			// Buscar misión activa del dron
			p.registry.mu.RLock()
			if drone, ok := p.registry.drones[msg.DroneID]; ok {
				telemetryLog.MissionID = drone.CurrentMission
			}
			p.registry.mu.RUnlock()
		}

		go func() {
			if err := database.DB.Create(&telemetryLog).Error; err != nil {
				log.Printf("Error guardando telemetría en BD: %v", err)
			}
		}()
	}

	// Reenviar TODO al GCS en tiempo real (no muestreado)
	p.hub.BroadcastToGCS(rawMap)
}

// handleMissionAck — Confirmación/rechazo de misión
func (p *ProtocolHandler) handleMissionAck(client *Client, raw []byte) {
	var msg models.MissionAckMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("Error parseando mission_ack: %v", err)
		return
	}

	// Actualizar misión en BD
	if database.DB != nil {
		status := msg.Status // "accepted" o "rejected"
		database.DB.Model(&models.Mission{}).Where("mission_id = ?", msg.MissionID).Update("status", status)
	}

	// Actualizar en memoria
	if msg.Status == "accepted" {
		p.registry.UpdateMissionStatus(msg.DroneID, msg.MissionID, "accepted")
	}

	// Notificar al GCS
	p.hub.BroadcastToGCS(map[string]interface{}{
		"type":      "mission_ack",
		"droneId":   msg.DroneID,
		"missionId": msg.MissionID,
		"status":    msg.Status,
	})

	log.Printf("✅ mission_ack — dron: %s, misión: %s, status: %s", msg.DroneID, msg.MissionID, msg.Status)
}

// handleMissionReady — Dron listo en punto de inicio
func (p *ProtocolHandler) handleMissionReady(client *Client, raw []byte) {
	var msg models.MissionReadyMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("Error parseando mission_ready: %v", err)
		return
	}

	// Actualizar misión en BD
	if database.DB != nil {
		database.DB.Model(&models.Mission{}).Where("mission_id = ?", msg.MissionID).Update("status", "in_progress")
	}

	// Actualizar en memoria
	p.registry.UpdateMissionStatus(msg.DroneID, msg.MissionID, "in_progress")

	// Notificar al GCS
	p.hub.BroadcastToGCS(map[string]interface{}{
		"type":      "mission_ready",
		"droneId":   msg.DroneID,
		"missionId": msg.MissionID,
	})

	log.Printf("🚀 mission_ready — dron: %s, misión: %s — TELEMETRÍA ACTIVA", msg.DroneID, msg.MissionID)
}

// handleMissionComplete — Misión finalizada
func (p *ProtocolHandler) handleMissionComplete(client *Client, raw []byte) {
	var msg models.MissionCompleteMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("Error parseando mission_complete: %v", err)
		return
	}

	// Actualizar misión en BD
	if database.DB != nil {
		database.DB.Model(&models.Mission{}).Where("mission_id = ?", msg.MissionID).Update("status", "completed")
		// Actualizar dron
		database.DB.Model(&models.Drone{}).Where("drone_id = ?", msg.DroneID).Update("status", "idle")
	}

	// Actualizar en memoria
	p.registry.CompleteMission(msg.DroneID, msg.MissionID)

	// Notificar al GCS
	p.hub.BroadcastToGCS(map[string]interface{}{
		"type":      "mission_complete",
		"droneId":   msg.DroneID,
		"missionId": msg.MissionID,
	})

	log.Printf("🏁 mission_complete — dron: %s, misión: %s", msg.DroneID, msg.MissionID)
}
