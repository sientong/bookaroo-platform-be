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
)

type PropertyHandlerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	handler *handlers.PropertyHandler
	router  *gin.Engine
}

func (suite *PropertyHandlerTestSuite) SetupSuite() {
	suite.db = tests.SetupTestDB(suite.T())
	suite.handler = handlers.NewPropertyHandler(suite.db)
	
	// Setup router
	suite.router = gin.New()
	suite.router.GET("/properties", suite.handler.ListProperties)
	suite.router.GET("/properties/:id", suite.handler.GetProperty)
	suite.router.GET("/properties/search", suite.handler.SearchProperties)
	suite.router.POST("/properties", suite.handler.CreateProperty)
	suite.router.PUT("/properties/:id", suite.handler.UpdateProperty)
	suite.router.GET("/properties/:id/owner-details", suite.handler.GetPropertyDetailsForOwner)
}

func (suite *PropertyHandlerTestSuite) SetupTest() {
	// Clear the database before each test
	suite.db.Exec("DELETE FROM property_images")
	suite.db.Exec("DELETE FROM properties")
	suite.db.Exec("DELETE FROM users")
}

func (suite *PropertyHandlerTestSuite) TestListProperties() {
	// Create test data
	owner := models.User{
		Email: "test@example.com",
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

	// Make request
	w := tests.MakeRequest(suite.router, "GET", "/properties", nil)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []models.Property
	tests.ParseResponse(suite.T(), w, &response)
	
	assert.Len(suite.T(), response, 1)
	assert.Equal(suite.T(), property.Name, response[0].Name)
}

func (suite *PropertyHandlerTestSuite) TestGetProperty() {
	// Create test data
	owner := models.User{
		Email: "test@example.com",
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

	// Make request
	w := tests.MakeRequest(suite.router, "GET", fmt.Sprintf("/properties/%d", property.ID), nil)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.Property
	tests.ParseResponse(suite.T(), w, &response)
	
	assert.Equal(suite.T(), property.Name, response.Name)
	assert.Equal(suite.T(), property.Description, response.Description)
}

func (suite *PropertyHandlerTestSuite) TestSearchProperties() {
	// Create test data
	owner := models.User{
		Email: "test@example.com",
		Name:  "Test Owner",
		Role:  "owner",
	}
	suite.db.Create(&owner)

	properties := []models.Property{
		{
			Name:     "Beach House",
			Location: "Miami",
			Price:    200.0,
			OwnerID:  owner.ID,
		},
		{
			Name:     "Mountain Cabin",
			Location: "Denver",
			Price:    150.0,
			OwnerID:  owner.ID,
		},
	}
	for _, p := range properties {
		suite.db.Create(&p)
	}

	// Test location search
	w := tests.MakeRequest(suite.router, "GET", "/properties/search?location=Miami", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []models.Property
	tests.ParseResponse(suite.T(), w, &response)
	
	assert.Len(suite.T(), response, 1)
	assert.Equal(suite.T(), "Beach House", response[0].Name)
}

func (suite *PropertyHandlerTestSuite) TestCreateProperty() {
	// Create test owner
	owner := models.User{
		Email: "owner@example.com",
		Name:  "Test Owner",
		Role:  "owner",
	}
	suite.db.Create(&owner)

	// Prepare request data
	propertyData := map[string]interface{}{
		"name":        "New Beach House",
		"description": "Beautiful beachfront property",
		"location":    "Bali",
		"price":       250.0,
		"amenities":   "WiFi, Pool, Beach Access",
		"owner_id":    owner.ID,
		"images": []map[string]interface{}{
			{
				"image_url": "https://example.com/image1.jpg",
			},
			{
				"image_url": "https://example.com/image2.jpg",
			},
		},
	}

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/properties", propertyData)

	// Assert response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.Property
	tests.ParseResponse(suite.T(), w, &response)

	// Verify property was created
	assert.Equal(suite.T(), propertyData["name"], response.Name)
	assert.Equal(suite.T(), propertyData["description"], response.Description)
	assert.Equal(suite.T(), propertyData["location"], response.Location)
	assert.Equal(suite.T(), propertyData["price"], response.Price)
	assert.Equal(suite.T(), propertyData["amenities"], response.Amenities)
	assert.Equal(suite.T(), owner.ID, response.OwnerID)
	assert.Len(suite.T(), response.Images, 2)

	// Verify images were created
	var images []models.PropertyImage
	suite.db.Where("property_id = ?", response.ID).Find(&images)
	assert.Len(suite.T(), images, 2)
}

func (suite *PropertyHandlerTestSuite) TestCreatePropertyInvalidOwner() {
	// Prepare request data with non-existent owner
	propertyData := map[string]interface{}{
		"name":        "New Beach House",
		"description": "Beautiful beachfront property",
		"location":    "Bali",
		"price":       250.0,
		"amenities":   "WiFi, Pool, Beach Access",
		"owner_id":    999, // Non-existent owner ID
	}

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/properties", propertyData)

	// Assert response
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *PropertyHandlerTestSuite) TestUpdateProperty() {
	// Create test owner
	owner := models.User{
		Email: "owner@example.com",
		Name:  "Test Owner",
		Role:  "owner",
	}
	suite.db.Create(&owner)

	// Create initial property
	property := models.Property{
		Name:        "Old Beach House",
		Description: "Old description",
		Location:    "Old location",
		Price:       200.0,
		Amenities:   "Old amenities",
		OwnerID:     owner.ID,
	}
	suite.db.Create(&property)

	// Create initial images
	oldImages := []models.PropertyImage{
		{PropertyID: property.ID, ImageURL: "https://example.com/old1.jpg"},
		{PropertyID: property.ID, ImageURL: "https://example.com/old2.jpg"},
	}
	for _, img := range oldImages {
		suite.db.Create(&img)
	}

	// Prepare update data
	updateData := map[string]interface{}{
		"name":        "Updated Beach House",
		"description": "Updated beachfront property",
		"location":    "Updated Bali",
		"price":       300.0,
		"amenities":   "Updated WiFi, Pool, Beach Access",
		"owner_id":    owner.ID,
		"images": []map[string]interface{}{
			{
				"image_url": "https://example.com/new1.jpg",
			},
			{
				"image_url": "https://example.com/new2.jpg",
			},
		},
	}

	// Make request
	w := tests.MakeRequest(suite.router, "PUT", fmt.Sprintf("/properties/%d", property.ID), updateData)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.Property
	tests.ParseResponse(suite.T(), w, &response)

	// Verify property was updated
	assert.Equal(suite.T(), updateData["name"], response.Name)
	assert.Equal(suite.T(), updateData["description"], response.Description)
	assert.Equal(suite.T(), updateData["location"], response.Location)
	assert.Equal(suite.T(), updateData["price"], response.Price)
	assert.Equal(suite.T(), updateData["amenities"], response.Amenities)
	assert.Equal(suite.T(), owner.ID, response.OwnerID)
	assert.Len(suite.T(), response.Images, 2)

	// Verify old images were replaced
	var images []models.PropertyImage
	suite.db.Where("property_id = ?", response.ID).Find(&images)
	assert.Len(suite.T(), images, 2)
	assert.Equal(suite.T(), "https://example.com/new1.jpg", images[0].ImageURL)
	assert.Equal(suite.T(), "https://example.com/new2.jpg", images[1].ImageURL)
}

func (suite *PropertyHandlerTestSuite) TestUpdatePropertyUnauthorized() {
	// Create owner 1
	owner1 := models.User{
		Email: "owner1@example.com",
		Name:  "Owner 1",
		Role:  "owner",
	}
	suite.db.Create(&owner1)

	// Create owner 2
	owner2 := models.User{
		Email: "owner2@example.com",
		Name:  "Owner 2",
		Role:  "owner",
	}
	suite.db.Create(&owner2)

	// Create property owned by owner1
	property := models.Property{
		Name:        "Beach House",
		Description: "Description",
		Location:    "Location",
		Price:       200.0,
		Amenities:   "Amenities",
		OwnerID:     owner1.ID,
	}
	suite.db.Create(&property)

	// Try to update property with owner2's ID
	updateData := map[string]interface{}{
		"name":        "Updated Beach House",
		"description": "Updated description",
		"location":    "Updated location",
		"price":       300.0,
		"owner_id":    owner2.ID,
	}

	// Make request
	w := tests.MakeRequest(suite.router, "PUT", fmt.Sprintf("/properties/%d", property.ID), updateData)

	// Assert response
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	// Verify property was not updated
	var checkProperty models.Property
	suite.db.First(&checkProperty, property.ID)
	assert.Equal(suite.T(), property.Name, checkProperty.Name)
	assert.Equal(suite.T(), owner1.ID, checkProperty.OwnerID)
}

func (suite *PropertyHandlerTestSuite) TestUpdateNonExistentProperty() {
	updateData := map[string]interface{}{
		"name":        "Updated Beach House",
		"description": "Updated description",
		"location":    "Updated location",
		"price":       300.0,
	}

	// Make request with non-existent property ID
	w := tests.MakeRequest(suite.router, "PUT", "/properties/999999", updateData)

	// Assert response
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *PropertyHandlerTestSuite) TestGetPropertyDetailsForOwner() {
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

	// Create property
	property := models.Property{
		Name:        "Beach House",
		Description: "Beautiful beachfront property",
		Location:    "Bali",
		Price:       200.0,
		Amenities:   "WiFi, Pool",
		OwnerID:     owner.ID,
	}
	suite.db.Create(&property)

	// Create some bookings
	now := time.Now()
	bookings := []models.Booking{
		{
			PropertyID: property.ID,
			UserID:     guest.ID,
			StartDate:  now.AddDate(0, 0, -10),  // Past booking
			EndDate:    now.AddDate(0, 0, -5),
			Status:     "completed",
			TotalPrice: 1000.0,
		},
		{
			PropertyID: property.ID,
			UserID:     guest.ID,
			StartDate:  now.AddDate(0, 0, 5),   // Future booking
			EndDate:    now.AddDate(0, 0, 10),
			Status:     "confirmed",
			TotalPrice: 1000.0,
		},
	}
	for _, booking := range bookings {
		suite.db.Create(&booking)
	}

	// Make request
	w := tests.MakeRequest(suite.router, "GET", fmt.Sprintf("/properties/%d/owner-details?owner_id=%d", property.ID, owner.ID), nil)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	tests.ParseResponse(suite.T(), w, &response)

	// Verify basic property info
	assert.Equal(suite.T(), property.Name, response["name"])
	assert.Equal(suite.T(), property.Description, response["description"])

	// Verify booking status
	assert.NotNil(suite.T(), response["is_available"])
	assert.NotNil(suite.T(), response["next_available_date"])

	// Verify booking history
	bookingHistory, ok := response["booking_history"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), bookingHistory, 2)

	// Verify booking statistics
	stats, ok := response["statistics"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.NotNil(suite.T(), stats["total_bookings"])
	assert.NotNil(suite.T(), stats["total_revenue"])
	assert.NotNil(suite.T(), stats["upcoming_bookings"])
}

func (suite *PropertyHandlerTestSuite) TestGetPropertyDetailsForOwnerUnauthorized() {
	// Create owner 1
	owner1 := models.User{
		Email: "owner1@example.com",
		Name:  "Owner 1",
		Role:  "owner",
	}
	suite.db.Create(&owner1)

	// Create owner 2
	owner2 := models.User{
		Email: "owner2@example.com",
		Name:  "Owner 2",
		Role:  "owner",
	}
	suite.db.Create(&owner2)

	// Create property owned by owner1
	property := models.Property{
		Name:        "Beach House",
		Description: "Description",
		Location:    "Location",
		Price:       200.0,
		Amenities:   "Amenities",
		OwnerID:     owner1.ID,
	}
	suite.db.Create(&property)

	// Try to get details with owner2's ID
	w := tests.MakeRequest(suite.router, "GET", fmt.Sprintf("/properties/%d/owner-details?owner_id=%d", property.ID, owner2.ID), nil)

	// Assert response
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *PropertyHandlerTestSuite) TestGetPropertyDetailsForOwnerNonExistent() {
	w := tests.MakeRequest(suite.router, "GET", "/properties/999999/owner-details?owner_id=1", nil)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func TestPropertyHandlerSuite(t *testing.T) {
	suite.Run(t, new(PropertyHandlerTestSuite))
}
