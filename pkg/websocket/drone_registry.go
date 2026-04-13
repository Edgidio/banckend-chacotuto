package websocket

import (
	"log"
	"sync"
	"time"

	"backend-chacotuto/app/models"
	"backend-chacotuto/pkg/database"
)

// ConnectedDrone es el estado en memoria de un dron conectado
type ConnectedDrone struct {
	DroneID       string
	Client        *Client
	Status        string // idle, mission_received, navigating, ready, in_mission
	LastHeartbeat time.Time
	LastTelemetry map[string]interface{} // Última telemetría completa (raw)
	CurrentMission string
	IsOnline      bool
}

// DroneRegistry mantiene el estado en memoria de todos los drones conectados
type DroneRegistry struct {
	mu     sync.RWMutex
	drones map[string]*ConnectedDrone
	hub    *Hub

	// Contro de muestreo para telemetría en BD
	lastSavedTelemetry map[string]time.Time
	saveMu             sync.Mutex
}

// NewDroneRegistry crea un nuevo registro
func NewDroneRegistry(hub *Hub) *DroneRegistry {
	return &DroneRegistry{
		drones:             make(map[string]*ConnectedDrone),
		hub:                hub,
		lastSavedTelemetry: make(map[string]time.Time),
	}
}

// Register agrega o actualiza un dron en el registro
func (r *DroneRegistry) Register(droneID string, client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.drones[droneID] = &ConnectedDrone{
		DroneID:       droneID,
		Client:        client,
		Status:        "idle",
		LastHeartbeat: time.Now(),
		IsOnline:      true,
	}

	client.DroneID = droneID

	log.Printf("📡 Dron registrado: %s", droneID)
}

// UpdateHeartbeat actualiza el timestamp y estado del heartbeat
func (r *DroneRegistry) UpdateHeartbeat(droneID, status string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if drone, ok := r.drones[droneID]; ok {
		drone.LastHeartbeat = time.Now()
		drone.Status = status
		drone.IsOnline = true
	}
}

// UpdateTelemetry actualiza la última telemetría del dron
func (r *DroneRegistry) UpdateTelemetry(droneID string, data map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if drone, ok := r.drones[droneID]; ok {
		drone.LastTelemetry = data
	}
}

// ShouldSaveTelemetry indica si han pasado >= 80ms para guardar toda la telemetría del dron
func (r *DroneRegistry) ShouldSaveTelemetry(droneID string) bool {
	r.saveMu.Lock()
	defer r.saveMu.Unlock()

	last, exists := r.lastSavedTelemetry[droneID]
	now := time.Now()

	// El dron manda cada 100ms. Si permitimos 80ms, aseguramos que pase sin ser filtrado.
	if !exists || now.Sub(last) >= 80*time.Millisecond {
		r.lastSavedTelemetry[droneID] = now
		return true
	}
	return false
}

// UpdateMissionStatus actualiza el estado de la misión actual del dron
func (r *DroneRegistry) UpdateMissionStatus(droneID, missionID, status string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if drone, ok := r.drones[droneID]; ok {
		drone.CurrentMission = missionID
		if status == "in_progress" {
			drone.Status = "in_mission"
		}
	}
}

// CompleteMission marca la misión como completada y el dron como idle
func (r *DroneRegistry) CompleteMission(droneID, missionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if drone, ok := r.drones[droneID]; ok {
		drone.CurrentMission = ""
		drone.Status = "idle"
	}
}

// MarkOffline marca un dron como desconectado
func (r *DroneRegistry) MarkOffline(droneID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if drone, ok := r.drones[droneID]; ok {
		drone.IsOnline = false
		drone.Status = "offline"
		drone.Client = nil
		
		// Sincronizar con BD
		if database.DB != nil {
			database.DB.Model(&models.Drone{}).Where("drone_id = ?", droneID).Updates(map[string]interface{}{
				"is_online": false,
				"status":    "offline",
			})
		}
		
		log.Printf("⚠️ Dron marcado offline (y en BD): %s", droneID)
	}
}

// GetAllDrones retorna una copia del estado de todos los drones
func (r *DroneRegistry) GetAllDrones() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(r.drones))
	for _, drone := range r.drones {
		d := map[string]interface{}{
			"droneId":       drone.DroneID,
			"status":        drone.Status,
			"lastHeartbeat": drone.LastHeartbeat,
			"isOnline":      drone.IsOnline,
		}
		if drone.LastTelemetry != nil {
			d["lastTelemetry"] = drone.LastTelemetry
		}
		if drone.CurrentMission != "" {
			d["currentMission"] = drone.CurrentMission
		}
		result = append(result, d)
	}
	return result
}

// GetDrone retorna el estado de un dron específico
func (r *DroneRegistry) GetDrone(droneID string) (map[string]interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	drone, ok := r.drones[droneID]
	if !ok {
		return nil, false
	}

	d := map[string]interface{}{
		"droneId":       drone.DroneID,
		"status":        drone.Status,
		"lastHeartbeat": drone.LastHeartbeat,
		"isOnline":      drone.IsOnline,
	}
	if drone.LastTelemetry != nil {
		d["lastTelemetry"] = drone.LastTelemetry
	}
	if drone.CurrentMission != "" {
		d["currentMission"] = drone.CurrentMission
	}

	return d, true
}

// IsDroneOnline verifica si un dron está conectado
func (r *DroneRegistry) IsDroneOnline(droneID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	drone, ok := r.drones[droneID]
	return ok && drone.IsOnline
}

// StartHeartbeatMonitor inicia una goroutine que revisa heartbeats cada 5 segundos
func (r *DroneRegistry) StartHeartbeatMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		r.checkHeartbeats()
	}
}

func (r *DroneRegistry) checkHeartbeats() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for droneID, drone := range r.drones {
		if drone.IsOnline && now.Sub(drone.LastHeartbeat) > 10*time.Second {
			drone.IsOnline = false
			drone.Status = "offline"
			drone.Client = nil

			// Sincronizar con BD
			if database.DB != nil {
				database.DB.Model(&models.Drone{}).Where("drone_id = ?", droneID).Updates(map[string]interface{}{
					"is_online": false,
					"status":    "offline",
				})
			}

			log.Printf("⏰ Dron %s timeout (sin heartbeat por >10s) — Actualizado en BD", droneID)

			// Notificar al GCS (en goroutine para no bloquear el lock)
			go r.hub.BroadcastToGCS(map[string]interface{}{
				"type":    "drone_offline",
				"droneId": droneID,
				"reason":  "heartbeat_timeout",
			})
		}
	}
}
