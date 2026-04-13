package database

import (
	"log"

	"backend-chacotuto/app/models"

	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// DB es la instancia global de la base de datos
var DB *gorm.DB

// Connect inicializa la conexión a SQLite y ejecuta las migraciones
func Connect() {
	var err error
	DB, err = gorm.Open(sqlite.Open("chacotuto.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error al conectar con la base de datos:", err)
	}

	log.Println("✅ Base de datos SQLite conectada correctamente")

	// Auto-migrar todas las tablas
	err = DB.AutoMigrate(
		&models.User{},
		&models.Drone{},
		&models.Mission{},
		&models.Waypoint{},
		&models.TelemetryLog{},
	)
	if err != nil {
		log.Fatal("Error al migrar tablas:", err)
	}

	log.Println("✅ Tablas migradas correctamente")

	// Seed: crear usuario admin si no existe
	seedAdmin()

	// Reset: marcar todos los drones como offline al iniciar el servidor
	ResetDroneStatus()
}

func ResetDroneStatus() {
	if DB == nil {
		return
	}
	log.Println("🔄 Reseteando estado de drones a offline...")
	DB.Model(&models.Drone{}).Where("1 = 1").Updates(map[string]interface{}{
		"is_online": false,
		"status":    "offline",
	})
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("chacotuto2026"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error generando hash del password:", err)
	}

	admin := models.User{
		Username: "admin",
		Password: string(hash),
		Role:     "admin",
	}

	if result := DB.Create(&admin); result.Error != nil {
		log.Fatal("Error creando usuario admin:", result.Error)
	}

	log.Println("✅ Usuario admin creado (admin / chacotuto2026)")
}
