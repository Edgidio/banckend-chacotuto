package main

import (
	"log"

	"backend-chacotuto/app/controllers"
	"backend-chacotuto/app/routes"
	"backend-chacotuto/pkg/database"
	"backend-chacotuto/pkg/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// 1. Inicializar la base de datos (SQLite + GORM)
	database.Connect()

	// 2. Inicializar la aplicación Fiber
	app := fiber.New(fiber.Config{
		AppName: "ChacOtuto Backend v1.0",
	})

	// 3. CORS — Permitir conexiones desde el frontend GCS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// 4. Instanciar el Hub de WebSockets
	hub := websocket.NewHub()
	go hub.Run()

	// 5. Instanciar los controladores con sus dependencias
	wsCtrl := controllers.NewWsController(hub)
	authCtrl := controllers.NewAuthController()
	droneCtrl := controllers.NewDroneController(hub)
	missionCtrl := controllers.NewMissionController(hub)
	statsCtrl := controllers.NewStatsController()

	// 6. Montar todas las rutas
	routes.SetupRoutes(app, wsCtrl, authCtrl, droneCtrl, missionCtrl, statsCtrl)

	// 7. Servir Frontend estático (Next.js Static Export con trailingSlash: true)
	// Con trailingSlash, cada ruta genera out/<ruta>/index.html
	// Fiber lo sirve nativamente al encontrar index.html en cada directorio.
	app.Static("/", "./out", fiber.Static{
		Index:  "index.html", // Sirve index.html cuando se accede a un directorio
		Browse: false,        // Sin listado de directorios
	})

	// Fallback: cualquier ruta sin archivo → 404 de Next.js
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).SendFile("./out/404.html")
	})

	// 8. Log de rutas disponibles
	log.Println("═══════════════════════════════════════════════")
	log.Println("  🚀 ChacOtuto Backend v1.0")
	log.Println("═══════════════════════════════════════════════")
	log.Println("  📡 WS Drones:  ws://localhost:8080/ws/drone")
	log.Println("  🎮 WS GCS:     ws://localhost:8080/ws/gcs?token=<JWT>")
	log.Println("  🔑 Login:      POST /api/auth/login")
	log.Println("  📋 Drones:     GET  /api/drones")
	log.Println("  🎯 Misiones:   GET  /api/missions")
	log.Println("  🎯 Crear:      POST /api/missions")
	log.Println("═══════════════════════════════════════════════")

	// 8. Iniciar la escucha
	if err := app.Listen(":8080"); err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}