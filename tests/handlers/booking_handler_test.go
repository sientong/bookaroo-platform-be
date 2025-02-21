package handlers_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/bookaroo/bookaroo-platform-be/handlers"
	"github.com/bookaroo/bookaroo-platform-be/models"
	"github.com/bookaroo/bookaroo-platform-be/tests"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"encoding/json"
)

type BookingHandlerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	handler *handlers.BookingHandler
	router  *gin.Engine
}

func (suite *BookingHandlerTestSuite) SetupSuite() {
	suite.db = tests.SetupTestDB(suite.T())
	suite.handler = handlers.NewBookingHandler(suite.db)
	
	// Setup router
	suite.router = gin.New()
	suite.router.POST("/bookings", suite.handler.CreateBooking)
	suite.router.GET("/bookings/guest/:guest_id", suite.handler.GetGuestBookings)
}

func (suite *BookingHandlerTestSuite) SetupTest() {
	// Clear the database before each test
	suite.db.Exec("DELETE FROM bookings")
	suite.db.Exec("DELETE FROM properties")
	suite.db.Exec("DELETE FROM users")
}

func (suite *BookingHandlerTestSuite) TestCreateBooking() {
	// Create test data
	owner := models.User{
		Email: "owner@example.com",
		Name:  "Test Owner",
		Role:  "owner",
	}
	suite.db.Create(&owner)

	guest := models.User{
		Email: "guest@example.com",
		Name:  "Test Guest",
		Role:  "guest",
	}
	suite.db.Create(&guest)

	property := models.Property{
		Name:        "Test Property",
		Description: "Test Description",
		Location:    "Test Location",
		Price:       100.0,
		OwnerID:     owner.ID,
	}
	suite.db.Create(&property)

	// Create booking request
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := startDate.AddDate(0, 0, 3)
	
	bookingRequest := handlers.CreateBookingRequest{
		PropertyID: property.ID,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/bookings", bookingRequest)

	// Assert response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.Booking
	tests.ParseResponse(suite.T(), w, &response)
	
	assert.Equal(suite.T(), property.ID, response.PropertyID)
	assert.Equal(suite.T(), "pending", response.Status)
	assert.Equal(suite.T(), 300.0, response.TotalPrice) // 3 days * 100.0 per day
}

func (suite *BookingHandlerTestSuite) TestCreateBookingConflict() {
	// Create test data
	owner := models.User{
		Email: "owner@example.com",
		Name:  "Test Owner",
		Role:  "owner",
	}
	suite.db.Create(&owner)

	property := models.Property{
		Name:        "Test Property",
		Description: "Test Description",
		Location:    "Test Location",
		Price:       100.0,
		OwnerID:     owner.ID,
	}
	suite.db.Create(&property)

	// Create existing booking
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := startDate.AddDate(0, 0, 3)
	
	existingBooking := models.Booking{
		PropertyID: property.ID,
		UserID:     1,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     "confirmed",
	}
	suite.db.Create(&existingBooking)

	// Try to create conflicting booking
	bookingRequest := handlers.CreateBookingRequest{
		PropertyID: property.ID,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/bookings", bookingRequest)

	// Assert response
	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

func (suite *BookingHandlerTestSuite) TestGetGuestBookings() {
	// Create test owner
	owner := models.User{
		Email: "owner@example.com",
		Name:  "Test Owner",
		Role:  "owner",
	}
	suite.db.Create(&owner)

	// Create test guest
	guest := models.User{
		Email: "guest@example.com",
		Name:  "Test Guest",
		Role:  "guest",
	}
	suite.db.Create(&guest)

	// Create test properties
	properties := []models.Property{
		{
			Name:        "Beach House",
			Description: "Beautiful beachfront property",
			Location:    "Bali",
			Price:       200.0,
			Amenities:   "WiFi, Pool",
			OwnerID:     owner.ID,
		},
		{
			Name:        "Mountain Cabin",
			Description: "Cozy mountain retreat",
			Location:    "Alps",
			Price:       150.0,
			Amenities:   "Fireplace, Hot Tub",
			OwnerID:     owner.ID,
		},
	}
	for _, property := range properties {
		suite.db.Create(&property)
	}

	// Create bookings for the guest
	now := time.Now()
	bookings := []models.Booking{
		{
			PropertyID: properties[0].ID,
			UserID:     guest.ID,
			StartDate:  now.AddDate(0, 0, -10), // Past booking
			EndDate:    now.AddDate(0, 0, -5),
			Status:     "completed",
			TotalPrice: 1000.0,
		},
		{
			PropertyID: properties[1].ID,
			UserID:     guest.ID,
			StartDate:  now.AddDate(0, 0, 5), // Future booking
			EndDate:    now.AddDate(0, 0, 10),
			Status:     "confirmed",
			TotalPrice: 750.0,
		},
	}
	for _, booking := range bookings {
		suite.db.Create(&booking)
	}

	// Make request
	w := tests.MakeRequest(suite.router, "GET", fmt.Sprintf("/bookings/guest/%d", guest.ID), nil)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	tests.ParseResponse(suite.T(), w, &response)

	// Verify bookings list
	bookingsList, ok := response["bookings"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), bookingsList, 2)

	// Verify booking details
	for _, booking := range bookingsList {
		bookingMap := booking.(map[string]interface{})
		assert.NotNil(suite.T(), bookingMap["id"])
		assert.NotNil(suite.T(), bookingMap["property"])
		assert.NotNil(suite.T(), bookingMap["start_date"])
		assert.NotNil(suite.T(), bookingMap["end_date"])
		assert.NotNil(suite.T(), bookingMap["status"])
		assert.NotNil(suite.T(), bookingMap["total_price"])
	}

	// Verify statistics
	stats, ok := response["statistics"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), float64(2), stats["total_bookings"])
	assert.Equal(suite.T(), float64(1750.0), stats["total_spent"])
	assert.Equal(suite.T(), float64(1), stats["upcoming_bookings"])
}

func (suite *BookingHandlerTestSuite) TestGetGuestBookingsNonExistentGuest() {
	w := tests.MakeRequest(suite.router, "GET", "/bookings/guest/999999", nil)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *BookingHandlerTestSuite) TestCreateBookingSuccess() {
    // First, log in to get a token
    token := suite.LoginAndGetToken()

    // Prepare booking request
    reqBody := map[string]interface{}{
        "property_id": 1,
        "start_date": "2024-03-16T00:00:00Z",
        "end_date": "2024-03-18T00:00:00Z",
    }
    jsonBody, _ := json.Marshal(reqBody)

    // Make request with token
    w := tests.MakeRequestWithToken(suite.router, "POST", "/bookings", jsonBody, token)

    // Assert response
    assert.Equal(suite.T(), http.StatusCreated, w.Code)
    var response map[string]interface{}
    tests.ParseResponse(suite.T(), w, &response)
    assert.Equal(suite.T(), "Booking created successfully", response["message"])
}

func (suite *BookingHandlerTestSuite) LoginAndGetToken() string {
    // Prepare login request
    loginBody := map[string]interface{}{
        "email":    "test@example.com",
        "password": "testpass123",
    }
    jsonBody, _ := json.Marshal(loginBody)

    // Make login request
    w := tests.MakeRequest(suite.router, "POST", "/login", jsonBody)
    var response map[string]interface{}
    tests.ParseResponse(suite.T(), w, &response)

    return response["token"].(string) // Return the token
}

func TestBookingHandlerSuite(t *testing.T) {
	suite.Run(t, new(BookingHandlerTestSuite))
}
