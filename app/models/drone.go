package models

import "time"

// Drone representa un dron registrado en la base de datos (persistente)
type Drone struct {
	DroneID       string    `gorm:"primaryKey;size:60" json:"droneId"` // CHACOTUTO-UUID
	Name          string    `gorm:"size:100" json:"name"`
	Status        string    `gorm:"size:30;default:offline" json:"status"` // offline, idle, navigating, ready, in_mission
	LastHeartbeat time.Time `json:"lastHeartbeat"`
	LastLat       float64   `json:"lastLat"`
	LastLng       float64   `json:"lastLng"`
	LastAlt       float64   `json:"lastAlt"`
	IsOnline      bool      `gorm:"default:false" json:"isOnline"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
