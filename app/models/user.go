package models

import "time"

// User representa un operador que puede acceder al GCS
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"` // bcrypt hash, nunca se serializa en JSON
	Role      string    `gorm:"size:20;default:operator" json:"role"` // admin, operator
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
