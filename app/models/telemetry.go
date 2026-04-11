package models

import "time"

// TelemetryLog almacena el historial de telemetría de cada dron por misión
type TelemetryLog struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	DroneID   string `gorm:"index;size:60" json:"droneId"`
	MissionID string `gorm:"index;size:60" json:"missionId"`
	Timestamp int64  `json:"timestamp"` // Unix ms del dron

	// Giroscopio (rad/s)
	GyroX float64 `json:"gyroX"`
	GyroY float64 `json:"gyroY"`
	GyroZ float64 `json:"gyroZ"`

	// Acelerómetro (m/s²)
	AccelX float64 `json:"accelX"`
	AccelY float64 `json:"accelY"`
	AccelZ float64 `json:"accelZ"`

	// Magnetómetro (μT)
	MagX float64 `json:"magX"`
	MagY float64 `json:"magY"`
	MagZ float64 `json:"magZ"`

	// Orientación (grados)
	Pitch float64 `json:"pitch"`
	Roll  float64 `json:"roll"`
	Yaw   float64 `json:"yaw"`

	// GPS
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`

	// Progreso de misión
	CurrentWaypoint int `json:"currentWaypoint"`
	TotalWaypoints  int `json:"totalWaypoints"`

	// Timestamp del servidor
	CreatedAt time.Time `json:"createdAt"`
}
