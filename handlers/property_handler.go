package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bookaroo/bookaroo-platform-be/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PropertyHandler struct {
	DB *gorm.DB
}

func NewPropertyHandler(db *gorm.DB) *PropertyHandler {
	return &PropertyHandler{DB: db}
}

// ListProperties returns all properties with optional filtering
func (h *PropertyHandler) ListProperties(c *gin.Context) {
	var properties []models.Property
	query := h.DB.Preload("Images").Preload("Owner")

	// Handle search parameters
	if location := c.Query("location"); location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}

	if err := query.Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching properties"})
		return
	}

	c.JSON(http.StatusOK, properties)
}

// GetProperty returns details of a specific property
func (h *PropertyHandler) GetProperty(c *gin.Context) {
	id := c.Param("id")
	var property models.Property

	if err := h.DB.Preload("Images").Preload("Owner").First(&property, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	c.JSON(http.StatusOK, property)
}

// SearchProperties handles property search with various filters
func (h *PropertyHandler) SearchProperties(c *gin.Context) {
	var properties []models.Property
	query := h.DB.Preload("Images").Preload("Owner")

	// Apply filters
	if location := c.Query("location"); location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			query = query.Where("price >= ?", price)
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			query = query.Where("price <= ?", price)
		}
	}

	if err := query.Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching properties"})
		return
	}

	c.JSON(http.StatusOK, properties)
}

type CreatePropertyRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description" binding:"required"`
	Location    string                   `json:"location" binding:"required"`
	Price       float64                  `json:"price" binding:"required"`
	Amenities   string                   `json:"amenities"`
	OwnerID     uint                     `json:"owner_id" binding:"required"`
	Images      []CreatePropertyImageRequest `json:"images"`
}

type CreatePropertyImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

type UpdatePropertyRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description" binding:"required"`
	Location    string                   `json:"location" binding:"required"`
	Price       float64                  `json:"price" binding:"required"`
	Amenities   string                   `json:"amenities"`
	OwnerID     uint                     `json:"owner_id" binding:"required"`
	Images      []CreatePropertyImageRequest `json:"images"`
}

// CreateProperty handles the creation of a new property
func (h *PropertyHandler) CreateProperty(c *gin.Context) {
	var req CreatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify owner exists and is an owner
	var owner models.User
	if err := h.DB.First(&owner, req.OwnerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Owner not found"})
		return
	}

	if owner.Role != "owner" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not an owner"})
		return
	}

	// Create property
	property := models.Property{
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		Price:       req.Price,
		Amenities:   req.Amenities,
		OwnerID:     req.OwnerID,
	}

	// Start a transaction
	tx := h.DB.Begin()

	if err := tx.Create(&property).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create property"})
		return
	}

	// Create property images
	for _, img := range req.Images {
		propertyImage := models.PropertyImage{
			PropertyID: property.ID,
			ImageURL:   img.ImageURL,
		}
		if err := tx.Create(&propertyImage).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create property images"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Load the created images
	h.DB.Preload("Images").First(&property, property.ID)

	c.JSON(http.StatusCreated, property)
}

// UpdateProperty handles updating an existing property
func (h *PropertyHandler) UpdateProperty(c *gin.Context) {
	var req UpdatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get property ID from URL
	propertyID := c.Param("id")

	// Check if property exists
	var existingProperty models.Property
	if err := h.DB.First(&existingProperty, propertyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	// Verify owner exists
	var owner models.User
	if err := h.DB.First(&owner, req.OwnerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Owner not found"})
		return
	}

	// Verify owner has permission to update this property
	if existingProperty.OwnerID != req.OwnerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this property"})
		return
	}

	// Start transaction
	tx := h.DB.Begin()

	// Update property details
	existingProperty.Name = req.Name
	existingProperty.Description = req.Description
	existingProperty.Location = req.Location
	existingProperty.Price = req.Price
	existingProperty.Amenities = req.Amenities

	if err := tx.Save(&existingProperty).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update property"})
		return
	}

	// Delete existing images
	if err := tx.Where("property_id = ?", existingProperty.ID).Delete(&models.PropertyImage{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update property images"})
		return
	}

	// Create new images
	for _, img := range req.Images {
		propertyImage := models.PropertyImage{
			PropertyID: existingProperty.ID,
			ImageURL:   img.ImageURL,
		}
		if err := tx.Create(&propertyImage).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new property images"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Load the updated property with images
	h.DB.Preload("Images").First(&existingProperty, existingProperty.ID)

	c.JSON(http.StatusOK, existingProperty)
}

type PropertyDetailsResponse struct {
	models.Property
	IsAvailable      bool           `json:"is_available"`
	NextAvailableDate *time.Time    `json:"next_available_date"`
	BookingHistory   []BookingInfo  `json:"booking_history"`
	Statistics       BookingStats   `json:"statistics"`
}

type BookingInfo struct {
	ID         uint      `json:"id"`
	GuestName  string    `json:"guest_name"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Status     string    `json:"status"`
	TotalPrice float64   `json:"total_price"`
}

type BookingStats struct {
	TotalBookings    int     `json:"total_bookings"`
	TotalRevenue     float64 `json:"total_revenue"`
	UpcomingBookings int     `json:"upcoming_bookings"`
}

// GetPropertyDetailsForOwner returns detailed property information for the owner
func (h *PropertyHandler) GetPropertyDetailsForOwner(c *gin.Context) {
	// Get property ID from URL
	propertyID := c.Param("id")

	// Get owner ID from query
	ownerID := c.Query("owner_id")
	if ownerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is required"})
		return
	}

	// Check if property exists
	var property models.Property
	if err := h.DB.Preload("Images").First(&property, propertyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	// Verify owner has permission to view this property
	if property.OwnerID != uint(property.OwnerID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this property's details"})
		return
	}

	// Get all bookings for this property
	var bookings []models.Booking
	h.DB.Preload("User").Where("property_id = ?", property.ID).Find(&bookings)

	// Prepare response
	response := PropertyDetailsResponse{
		Property: property,
	}

	now := time.Now()
	response.IsAvailable = true
	var nextAvailable time.Time

	// Calculate booking statistics and check availability
	stats := BookingStats{}
	bookingHistory := make([]BookingInfo, 0)

	for _, booking := range bookings {
		// Add to booking history
		bookingInfo := BookingInfo{
			ID:         booking.ID,
			GuestName:  booking.User.Name,
			StartDate:  booking.StartDate,
			EndDate:    booking.EndDate,
			Status:     booking.Status,
			TotalPrice: booking.TotalPrice,
		}
		bookingHistory = append(bookingHistory, bookingInfo)

		// Update statistics
		stats.TotalBookings++
		stats.TotalRevenue += booking.TotalPrice

		// Check if booking affects current availability
		if booking.Status == "confirmed" || booking.Status == "pending" {
			if booking.StartDate.Before(now) && booking.EndDate.After(now) {
				response.IsAvailable = false
				if nextAvailable.IsZero() || booking.EndDate.After(nextAvailable) {
					nextAvailable = booking.EndDate
				}
			}
			if booking.StartDate.After(now) {
				stats.UpcomingBookings++
			}
		}
	}

	if !response.IsAvailable && !nextAvailable.IsZero() {
		response.NextAvailableDate = &nextAvailable
	}

	response.BookingHistory = bookingHistory
	response.Statistics = stats

	c.JSON(http.StatusOK, response)
}
