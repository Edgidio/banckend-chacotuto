package routes

import (
	"backend-chacotuto/app/controllers"
	"backend-chacotuto/app/middleware"

	"github.com/gofiber/fiber/v2"
	fiberWebsocket "github.com/gofiber/websocket/v2"
)

// SetupRoutes vincula cada endpoint de URL con sus respectivos controladores
func SetupRoutes(
	app *fiber.App,
	wsCtrl *controllers.WsController,
	authCtrl *controllers.AuthController,
	droneCtrl *controllers.DroneController,
	missionCtrl *controllers.MissionController,
	statsCtrl *controllers.StatsController,
) {

	// ===================================================================
	// Rutas públicas (sin autenticación)
	// ===================================================================

	// Auth
	auth := app.Group("/api/auth")
	auth.Post("/login", authCtrl.Login)

	// ===================================================================
	// WebSocket para drones — SIN autenticación JWT
	// Los drones se identifican con su droneId en el mensaje "register"
	// ===================================================================

	app.Use("/ws/drone", wsCtrl.UpgradeDroneMiddleware)
	app.Get("/ws/drone", fiberWebsocket.New(wsCtrl.ServeWSDrone))

	// ===================================================================
	// Rutas protegidas (requieren JWT)
	// ===================================================================

	// Grupo API protegido
	api := app.Group("/api", middleware.AuthRequired())

	// Analíticas / Dashboard
	api.Get("/stats", statsCtrl.GetStats)

	// Auth protegida
	api.Post("/auth/register", authCtrl.Register) // Solo admin puede crear usuarios
	api.Get("/auth/me", authCtrl.Me)

	// Drones
	api.Get("/drones", droneCtrl.GetAll)
	api.Get("/drones/:id", droneCtrl.GetOne)
	api.Get("/drones/:id/telemetry", droneCtrl.GetTelemetry)

	// Misiones
	api.Get("/missions", missionCtrl.GetAll)
	api.Get("/missions/:id", missionCtrl.GetOne)
	api.Post("/missions", missionCtrl.Create)
	api.Delete("/missions/:id", missionCtrl.Cancel)

	// ===================================================================
	// WebSocket para GCS — CON autenticación JWT (token en query param)
	// ws://host:8080/ws/gcs?token=<JWT>
	// ===================================================================

	app.Use("/ws/gcs", wsCtrl.UpgradeGCSMiddleware)
	app.Get("/ws/gcs", fiberWebsocket.New(wsCtrl.ServeWSGCS))
}
