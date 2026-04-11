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

	// 6. Montar todas las rutas
	routes.SetupRoutes(app, wsCtrl, authCtrl, droneCtrl, missionCtrl)

	// 7. Servir Frontend estático (Next.js SPA)
	// Servir archivos directos desde la carpeta ./out
	app.Static("/", "./out")

	// Capturar cualquier ruta restante y servir index.html (SPA fallback)
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./out/index.html")
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