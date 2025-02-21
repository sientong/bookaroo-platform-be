package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/bookaroo/bookaroo-platform-be/handlers"
	"github.com/bookaroo/bookaroo-platform-be/models"
	"github.com/bookaroo/bookaroo-platform-be/routes"
	"github.com/bookaroo/bookaroo-platform-be/tests"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type APIIntegrationTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
}

func (suite *APIIntegrationTestSuite) SetupSuite() {
	suite.db = tests.SetupTestDB(suite.T())

	// Setup router with all routes
	suite.router = gin.New()
	routes.SetupRoutes(suite.router, suite.db)
}

func (suite *APIIntegrationTestSuite) SetupTest() {
	// Clear the database before each test
	suite.db.Exec("DELETE FROM bookings")
	suite.db.Exec("DELETE FROM property_images")
	suite.db.Exec("DELETE FROM properties")
	suite.db.Exec("DELETE FROM users")
}

func (suite *APIIntegrationTestSuite) TestFullBookingFlow() {
	// 1. Create owner and property
	owner := models.User{
		Email: "owner@example.com",
		Name:  "Test Owner",
		Role:  "owner",
	}
	suite.db.Create(&owner)

	property := models.Property{
		Name:        "Luxury Villa",
		Description: "Beautiful villa with ocean view",
		Location:    "Bali",
		Price:       200.0,
		OwnerID:     owner.ID,
	}
	suite.db.Create(&property)

	// 2. Search for properties in Bali
	w := tests.MakeRequest(suite.router, "GET", "/api/properties/search?location=Bali", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var searchResponse []models.Property
	tests.ParseResponse(suite.T(), w, &searchResponse)
	assert.Len(suite.T(), searchResponse, 1)
	assert.Equal(suite.T(), "Luxury Villa", searchResponse[0].Name)

	// 3. Get property details
	w = tests.MakeRequest(suite.router, "GET", "/api/properties/"+fmt.Sprint(property.ID), nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var propertyResponse models.Property
	tests.ParseResponse(suite.T(), w, &propertyResponse)
	assert.Equal(suite.T(), property.Name, propertyResponse.Name)

	// 4. Create a booking
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := startDate.AddDate(0, 0, 3)

	bookingRequest := handlers.CreateBookingRequest{
		PropertyID: property.ID,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	w = tests.MakeRequest(suite.router, "POST", "/api/bookings", bookingRequest)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var bookingResponse models.Booking
	tests.ParseResponse(suite.T(), w, &bookingResponse)
	assert.Equal(suite.T(), property.ID, bookingResponse.PropertyID)
	assert.Equal(suite.T(), "pending", bookingResponse.Status)

	// 5. Check owner's dashboard
	w = tests.MakeRequest(suite.router, "GET", "/api/dashboard", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var dashboardResponse map[string]interface{}
	tests.ParseResponse(suite.T(), w, &dashboardResponse)
	assert.Equal(suite.T(), "owner", dashboardResponse["role"])
}

func (suite *APIIntegrationTestSuite) TestPropertyAvailabilityFlow() {
	// 1. Create owner and property
	owner := models.User{
		Email: "owner@example.com",
		Name:  "Test Owner",
		Role:  "owner",
	}
	suite.db.Create(&owner)

	property := models.Property{
		Name:        "Beach House",
		Description: "Cozy beach house",
		Location:    "Miami",
		Price:       150.0,
		OwnerID:     owner.ID,
	}
	suite.db.Create(&property)

	// 2. Create first booking
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := startDate.AddDate(0, 0, 3)

	firstBooking := handlers.CreateBookingRequest{
		PropertyID: property.ID,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	w := tests.MakeRequest(suite.router, "POST", "/api/bookings", firstBooking)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// 3. Try to create overlapping booking
	secondBooking := handlers.CreateBookingRequest{
		PropertyID: property.ID,
		StartDate:  startDate.AddDate(0, 0, 1),
		EndDate:    endDate.AddDate(0, 0, 1),
	}

	w = tests.MakeRequest(suite.router, "POST", "/api/bookings", secondBooking)
	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

func TestAPIIntegrationSuite(t *testing.T) {
	suite.Run(t, new(APIIntegrationTestSuite))
}
