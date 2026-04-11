package controllers

import (
	"backend-chacotuto/app/middleware"
	"backend-chacotuto/app/models"
	"backend-chacotuto/pkg/database"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// AuthController maneja login, register y perfil
type AuthController struct{}

// NewAuthController crea la instancia del controlador
func NewAuthController() *AuthController {
	return &AuthController{}
}

// LoginRequest es el cuerpo de la petición de login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest es el cuerpo de la petición de registro
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Login autentica un usuario y retorna un JWT
func (a *AuthController) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Datos de login inválidos",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username y password son requeridos",
		})
	}

	// Buscar usuario en la BD
	var user models.User
	result := database.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Credenciales inválidas",
		})
	}

	// Comparar password con bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Credenciales inválidas",
		})
	}

	// Generar JWT
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generando token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// Register crea un nuevo usuario (solo admin puede crear)
func (a *AuthController) Register(c *fiber.Ctx) error {
	// Verificar que el solicitante es admin
	role, _ := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Solo administradores pueden crear usuarios",
		})
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Datos de registro inválidos",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username y password son requeridos",
		})
	}

	if req.Role == "" {
		req.Role = "operator"
	}

	// Hash del password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error procesando password",
		})
	}

	user := models.User{
		Username: req.Username,
		Password: string(hash),
		Role:     req.Role,
	}

	if result := database.DB.Create(&user); result.Error != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "El usuario ya existe",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Usuario creado exitosamente",
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// Me retorna los datos del usuario autenticado
func (a *AuthController) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals("userId").(uint)
	username, _ := c.Locals("username").(string)
	role, _ := c.Locals("role").(string)

	return c.JSON(fiber.Map{
		"id":       userID,
		"username": username,
		"role":     role,
	})
}
