package models

// ===================================================================
// Modelos del protocolo WebSocket (CHACOTUTO_PROTOCOLO_WEBSOCKET.md)
// ===================================================================

// --- Estructuras base para mensajes genéricos ---

// BaseMessage contiene los campos comunes de todo mensaje del protocolo
type BaseMessage struct {
	Type      string `json:"type"`
	DroneID   string `json:"droneId,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// --- Mensajes: App → Backend ---

// RegisterMsg — Registro del dron al conectarse
type RegisterMsg struct {
	Type      string `json:"type"`
	DroneID   string `json:"droneId"`
	Timestamp int64  `json:"timestamp"`
}

// HeartbeatMsg — Latido periódico cada 2 segundos
type HeartbeatMsg struct {
	Type      string `json:"type"`
	DroneID   string `json:"droneId"`
	Status    string `json:"status"` // idle, mission_received, navigating, ready, in_mission
	Timestamp int64  `json:"timestamp"`
}

// TelemetryMsg — Datos de sensores en tiempo real (cada 100ms durante misión)
type TelemetryMsg struct {
	Type        string             `json:"type"`
	DroneID     string             `json:"droneId"`
	Timestamp   int64              `json:"timestamp"`
	Sensors     SensorPayload      `json:"sensors"`
	Orientation OrientationPayload `json:"orientation"`
	GPS         *GPSPayload        `json:"gps,omitempty"`
	Mission     *MissionProgress   `json:"mission,omitempty"`
	Battery     *BatteryPayload    `json:"battery,omitempty"`
}

// BatteryPayload contiene el estado de la batería del dispositivo
type BatteryPayload struct {
	Level      int  `json:"level"`
	IsCharging bool `json:"isCharging"`
}

// SensorPayload agrupa los tres sensores del dispositivo
type SensorPayload struct {
	Gyroscope     Vec3 `json:"gyroscope"`
	Accelerometer Vec3 `json:"accelerometer"`
	Magnetometer  Vec3 `json:"magnetometer"`
}

// Vec3 es un vector tridimensional genérico
type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// OrientationPayload contiene la orientación fusionada del dispositivo
type OrientationPayload struct {
	Pitch float64 `json:"pitch"` // -180 a 180 grados
	Roll  float64 `json:"roll"`  // -90 a 90 grados
	Yaw   float64 `json:"yaw"`   // 0 a 360 grados (heading)
}

// GPSPayload contiene la posición GPS (puede ser null si sin señal)
type GPSPayload struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
}

// MissionProgress indica el progreso durante una misión activa
type MissionProgress struct {
	Status              string `json:"status"` // in_progress
	CurrentWaypointIndex int   `json:"currentWaypointIndex"`
	TotalWaypoints      int    `json:"totalWaypoints"`
}

// MissionAckMsg — Confirmación/rechazo de misión recibida
type MissionAckMsg struct {
	Type      string `json:"type"`
	DroneID   string `json:"droneId"`
	MissionID string `json:"missionId"`
	Status    string `json:"status"` // accepted, rejected
	Timestamp int64  `json:"timestamp"`
}

// MissionReadyMsg — Dron listo en punto de inicio
type MissionReadyMsg struct {
	Type      string `json:"type"`
	DroneID   string `json:"droneId"`
	MissionID string `json:"missionId"`
	Timestamp int64  `json:"timestamp"`
}

// MissionCompleteMsg — Misión finalizada
type MissionCompleteMsg struct {
	Type      string `json:"type"`
	DroneID   string `json:"droneId"`
	MissionID string `json:"missionId"`
	Timestamp int64  `json:"timestamp"`
}

// --- Mensajes: Backend → App ---

// MissionAssignMsg — Asignar misión al dron
type MissionAssignMsg struct {
	Type       string        `json:"type"`
	MissionID  string        `json:"missionId"`
	Waypoints  []WaypointMsg `json:"waypoints"`
	StartPoint *WaypointMsg  `json:"startPoint,omitempty"`
}

// WaypointMsg es un punto de ruta en un mensaje WebSocket
type WaypointMsg struct {
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Alt    float64 `json:"alt"`
	Action string  `json:"action"` // takeoff, waypoint, land, start
	Index  int     `json:"index"`
}

// MissionCancelMsg — Cancelar misión
type MissionCancelMsg struct {
	Type      string `json:"type"`
	MissionID string `json:"missionId"`
	Reason    string `json:"reason"`
}

// CommandMsg — Comando genérico extensible
type CommandMsg struct {
	Type    string         `json:"type"`
	Command string         `json:"command"`
	Params  map[string]any `json:"params,omitempty"`
}

// --- Mensajes internos: Backend → GCS ---

// GCSDroneEvent notifica al GCS sobre cambios en el estado de un dron
type GCSDroneEvent struct {
	Type    string `json:"type"` // drone_online, drone_offline, drone_status
	DroneID string `json:"droneId"`
	Status  string `json:"status,omitempty"`
}
