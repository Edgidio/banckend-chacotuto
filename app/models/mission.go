package models

import "time"

// Mission representa una misión asignada a un dron
type Mission struct {
	MissionID string     `gorm:"primaryKey;size:60" json:"missionId"`
	DroneID   string     `gorm:"index;size:60" json:"droneId"`
	Status    string     `gorm:"size:20;default:pending" json:"status"` // pending, accepted, rejected, in_progress, completed, cancelled
	Waypoints []Waypoint `gorm:"foreignKey:MissionID;references:MissionID" json:"waypoints"`
	StartLat  float64    `json:"startLat"`
	StartLng  float64    `json:"startLng"`
	StartAlt  float64    `json:"startAlt"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// Waypoint es un punto de ruta dentro de una misión
type Waypoint struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	MissionID string  `gorm:"index;size:60" json:"missionId"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Alt       float64 `json:"alt"`
	Action    string  `gorm:"size:20" json:"action"` // takeoff, waypoint, land, start
	Index     int     `json:"index"`
}
