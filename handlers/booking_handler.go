package handlers

import (
	"net/http"
	"time"

	"github.com/bookaroo/bookaroo-platform-be/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookingHandler struct {
	DB *gorm.DB
}

func NewBookingHandler(db *gorm.DB) *BookingHandler {
	return &BookingHandler{DB: db}
}

type CreateBookingRequest struct {
	PropertyID uint      `json:"property_id" binding:"required"`
	StartDate  time.Time `json:"start_date" binding:"required"`
	EndDate    time.Time `json:"end_date" binding:"required"`
}

type GuestBookingResponse struct {
	ID         uint            `json:"id"`
	Property   PropertyDetails `json:"property"`
	StartDate  time.Time       `json:"start_date"`
	EndDate    time.Time       `json:"end_date"`
	Status     string          `json:"status"`
	TotalPrice float64         `json:"total_price"`
}

type PropertyDetails struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	Price       float64 `json:"price"`
	Amenities   string  `json:"amenities"`
}

type GuestBookingsResponse struct {
	Bookings   []GuestBookingResponse `json:"bookings"`
	Statistics GuestBookingStats      `json:"statistics"`
}

type GuestBookingStats struct {
	TotalBookings    int     `json:"total_bookings"`
	TotalSpent       float64 `json:"total_spent"`
	UpcomingBookings int     `json:"upcoming_bookings"`
}

// CreateBooking handles new booking creation
// @Summary Create a new booking
// @Description Create a new booking with the given details
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body CreateBookingRequest true "Booking details"
// @Success 201 {object} models.Booking
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID, _ := c.Get("user_id") // Get user ID from context

	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if property exists
	var property models.Property
	if err := h.DB.First(&property, req.PropertyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	// Check if dates are available
	var conflictingBookings int64
	h.DB.Model(&models.Booking{}).
		Where("property_id = ? AND status != 'cancelled' AND "+
			"((start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?))",
			req.PropertyID, req.StartDate, req.EndDate, req.StartDate, req.EndDate).
		Count(&conflictingBookings)

	if conflictingBookings > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Property is not available for these dates"})
		return
	}

	// Calculate total price
	days := req.EndDate.Sub(req.StartDate).Hours() / 24
	totalPrice := property.Price * float64(days)

	booking := models.Booking{
		PropertyID: req.PropertyID,
		UserID:     userID.(uint),
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		TotalPrice: totalPrice,
		Status:     "pending",
	}

	if err := h.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Booking created successfully"})
}

// GetGuestBookings returns a list of bookings for a specific guest
// @Summary Get bookings for a guest
// @Description Retrieve a list of bookings for the authenticated guest
// @Tags bookings
// @Accept json
// @Produce json
// @Success 200 {object} GuestBookingsResponse
// @Failure 401 {object} map[string]string
// @Router /bookings [get]
func (h *BookingHandler) GetGuestBookings(c *gin.Context) {
	// Get guest ID from URL
	guestID := c.Param("guest_id")

	// Check if guest exists
	var guest models.User
	if err := h.DB.First(&guest, guestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guest not found"})
		return
	}

	// Get all bookings for this guest
	var bookings []models.Booking
	if err := h.DB.Preload("Property").Where("user_id = ?", guestID).Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}

	// Prepare response
	response := GuestBookingsResponse{
		Bookings:   make([]GuestBookingResponse, 0),
		Statistics: GuestBookingStats{},
	}

	now := time.Now()

	// Process bookings
	for _, booking := range bookings {
		// Add to bookings list
		bookingResponse := GuestBookingResponse{
			ID: booking.ID,
			Property: PropertyDetails{
				ID:          booking.Property.ID,
				Name:        booking.Property.Name,
				Description: booking.Property.Description,
				Location:    booking.Property.Location,
				Price:       booking.Property.Price,
				Amenities:   booking.Property.Amenities,
			},
			StartDate:  booking.StartDate,
			EndDate:    booking.EndDate,
			Status:     booking.Status,
			TotalPrice: booking.TotalPrice,
		}
		response.Bookings = append(response.Bookings, bookingResponse)

		// Update statistics
		response.Statistics.TotalBookings++
		response.Statistics.TotalSpent += booking.TotalPrice

		// Count upcoming bookings
		if booking.StartDate.After(now) && (booking.Status == "confirmed" || booking.Status == "pending") {
			response.Statistics.UpcomingBookings++
		}
	}

	c.JSON(http.StatusOK, response)
}
