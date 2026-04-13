package controllers

import (
	"backend-chacotuto/app/models"
	"backend-chacotuto/pkg/database"

	"github.com/gofiber/fiber/v2"
)

// StatsController maneja las analíticas de todo el sistema
type StatsController struct{}

// NewStatsController crea la instancia
func NewStatsController() *StatsController {
	return &StatsController{}
}

// GetStats consolida las estadísticas para el dashboard general
func (sc *StatsController) GetStats(c *fiber.Ctx) error {
	var totalDrones int64
	var onlineDrones int64

	// Conteo de drones
	database.DB.Model(&models.Drone{}).Count(&totalDrones)
	database.DB.Model(&models.Drone{}).Where("is_online = ?", true).Count(&onlineDrones)

	// Conteo de Misiones por estado
	var totalMissions int64
	database.DB.Model(&models.Mission{}).Count(&totalMissions)

	var activeMissions int64
	database.DB.Model(&models.Mission{}).Where("status IN ?", []string{"pending", "accepted", "in_progress", "ready", "navigating"}).Count(&activeMissions)

	var completedMissions int64
	database.DB.Model(&models.Mission{}).Where("status = ?", "completed").Count(&completedMissions)

	var failedMissions int64
	database.DB.Model(&models.Mission{}).Where("status IN ?", []string{"rejected", "cancelled"}).Count(&failedMissions)

	// Misiones recientes
	var recentMissions []models.Mission
	database.DB.Order("created_at DESC").Limit(5).Find(&recentMissions)

	// Opcional: Tiempo de vuelo/telemetría total (por ahora vamos a contar logs como "puntos de telemetría procesados")
	var telemetryPoints int64
	database.DB.Model(&models.TelemetryLog{}).Count(&telemetryPoints)

	// Cálculo de salud de la flota
	fleetHealth := 100.0
	if totalDrones > 0 {
		fleetHealth = float64(onlineDrones) / float64(totalDrones) * 100.0
	}

	return c.JSON(fiber.Map{
		"drones": fiber.Map{
			"total":  totalDrones,
			"online": onlineDrones,
			"health": fleetHealth,
		},
		"missions": fiber.Map{
			"total":     totalMissions,
			"active":    activeMissions,
			"completed": completedMissions,
			"failed":    failedMissions,
		},
		"telemetry": fiber.Map{
			"dataPoints": telemetryPoints,
		},
		"recentActivity": recentMissions,
	})
}
