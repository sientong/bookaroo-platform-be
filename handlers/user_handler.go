package handlers

import (
	"net/http"
	"regexp"

	"github.com/bookaroo/bookaroo-platform-be/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/bookaroo/bookaroo-platform-be/middleware"
)

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

type RegisterOwnerRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	Name         string `json:"name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Address      string `json:"address" binding:"required"`
	BusinessName string `json:"business_name" binding:"required"`
}

type RegisterGuestRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Address  string `json:"address" binding:"required"`
}

type RegisterResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// GetUserDashboard returns dashboard data based on user role
func (h *UserHandler) GetUserDashboard(c *gin.Context) {
	userID := uint(1) // TODO: Replace with actual user ID from context
	var user models.User

	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.Role == "owner" {
		// Get owner's properties and their bookings
		var properties []models.Property
		if err := h.DB.Preload("Bookings").Where("owner_id = ?", userID).Find(&properties).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching properties"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"role":       "owner",
			"properties": properties,
		})
	} else {
		// Get guest's bookings
		var bookings []models.Booking
		if err := h.DB.Preload("Property").Where("user_id = ?", userID).Find(&bookings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching bookings"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"role":     "guest",
			"bookings": bookings,
		})
	}
}

// RegisterOwner handles the registration of a new property owner
func (h *UserHandler) RegisterOwner(c *gin.Context) {
	var req RegisterOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		Password:     string(hashedPassword),
		Name:         req.Name,
		Role:         "owner",
		Phone:        req.Phone,
		Address:      req.Address,
		BusinessName: &req.BusinessName,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Prepare response
	response := RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}

	c.JSON(http.StatusCreated, response)
}

// RegisterGuest handles the registration of a new guest user
func (h *UserHandler) RegisterGuest(c *gin.Context) {
	var req RegisterGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user
	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		Role:     "guest",
		Phone:    req.Phone,
		Address:  req.Address,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Prepare response
	response := RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}

	c.JSON(http.StatusCreated, response)
}

// Login authenticates a user and returns a JWT token
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	response := LoginResponse{
		Token: token,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
