package handlers

import (
	"net/http"

	"github.com/bookaroo/bookaroo-platform-be/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
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
