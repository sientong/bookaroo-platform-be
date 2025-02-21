package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bookaroo/bookaroo-platform-be/handlers"
	"github.com/bookaroo/bookaroo-platform-be/models"
	"github.com/bookaroo/bookaroo-platform-be/tests"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandlerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	handler *handlers.UserHandler
	router  *gin.Engine
}

func (suite *UserHandlerTestSuite) SetupSuite() {
	suite.db = tests.SetupTestDB(suite.T())
	suite.handler = handlers.NewUserHandler(suite.db)
	
	// Setup router
	suite.router = gin.New()
	suite.router.POST("/register/owner", suite.handler.RegisterOwner)
	suite.router.POST("/register/guest", suite.handler.RegisterGuest)
	suite.router.POST("/login", suite.handler.Login)
}

func (suite *UserHandlerTestSuite) SetupTest() {
	tests.ClearTestDB(suite.T(), suite.db)
}

func (suite *UserHandlerTestSuite) TestRegisterOwner() {
	// Prepare request body
	reqBody := map[string]interface{}{
		"email":     "owner@example.com",
		"password":  "securepass123",
		"name":      "Test Owner",
		"phone":     "+1234567890",
		"address":   "123 Property St",
		"business_name": "Luxury Rentals",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/register/owner", bytes.NewBuffer(jsonBody))

	// Assert response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	tests.ParseResponse(suite.T(), w, &response)

	// Verify user was created
	var user models.User
	err := suite.db.Where("email = ?", reqBody["email"]).First(&user).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), reqBody["email"], user.Email)
	assert.Equal(suite.T(), reqBody["name"], user.Name)
	assert.Equal(suite.T(), "owner", user.Role)

	// Verify password was hashed
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqBody["password"].(string)))
	assert.NoError(suite.T(), err)
}

func (suite *UserHandlerTestSuite) TestRegisterGuest() {
	// Prepare request body
	reqBody := map[string]interface{}{
		"email":    "guest@example.com",
		"password": "securepass123",
		"name":     "Test Guest",
		"phone":    "+1234567890",
		"address":  "456 Guest Ave",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/register/guest", bytes.NewBuffer(jsonBody))

	// Assert response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	tests.ParseResponse(suite.T(), w, &response)

	// Verify user was created
	var user models.User
	err := suite.db.Where("email = ?", reqBody["email"]).First(&user).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), reqBody["email"], user.Email)
	assert.Equal(suite.T(), reqBody["name"], user.Name)
	assert.Equal(suite.T(), "guest", user.Role)

	// Verify password was hashed
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqBody["password"].(string)))
	assert.NoError(suite.T(), err)
}

func (suite *UserHandlerTestSuite) TestRegisterOwnerDuplicateEmail() {
	// Create existing user
	existingUser := models.User{
		Email:    "owner@example.com",
		Password: "hashedpass",
		Name:     "Existing Owner",
		Role:     "owner",
	}
	suite.db.Create(&existingUser)

	// Try to register with same email
	reqBody := map[string]interface{}{
		"email":     "owner@example.com",
		"password":  "securepass123",
		"name":      "Test Owner",
		"phone":     "+1234567890",
		"address":   "123 Property St",
		"business_name": "Luxury Rentals",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/register/owner", bytes.NewBuffer(jsonBody))

	// Assert response
	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

func (suite *UserHandlerTestSuite) TestRegisterGuestInvalidEmail() {
	// Prepare request body with invalid email
	reqBody := map[string]interface{}{
		"email":    "invalid-email",
		"password": "securepass123",
		"name":     "Test Guest",
		"phone":    "+1234567890",
		"address":  "456 Guest Ave",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/register/guest", bytes.NewBuffer(jsonBody))

	// Assert response
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *UserHandlerTestSuite) TestLoginSuccess() {
	// Create a test user with hashed password
	password := "testpass123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Name:     "Test User",
		Role:     "guest",
	}
	suite.db.Create(&user)

	// Prepare login request
	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "testpass123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := tests.MakeRequest(suite.router, "POST", "/login", bytes.NewBuffer(jsonBody))

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	tests.ParseResponse(suite.T(), w, &response)

	// Verify response contains token and user info
	assert.NotEmpty(suite.T(), response["token"])
	assert.Equal(suite.T(), user.Email, response["email"])
	assert.Equal(suite.T(), user.Role, response["role"])
}

func (suite *UserHandlerTestSuite) TestLoginInvalidCredentials() {
	// Create a test user
	password := "testpass123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Name:     "Test User",
		Role:     "guest",
	}
	suite.db.Create(&user)

	// Test cases for invalid credentials
	testCases := []struct {
		email    string
		password string
	}{
		{"test@example.com", "wrongpassword"},
		{"wrong@example.com", "testpass123"},
		{"", "testpass123"},
		{"test@example.com", ""},
	}

	for _, tc := range testCases {
		reqBody := map[string]interface{}{
			"email":    tc.email,
			"password": tc.password,
		}
		jsonBody, _ := json.Marshal(reqBody)

		w := tests.MakeRequest(suite.router, "POST", "/login", bytes.NewBuffer(jsonBody))
		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	}
}

func TestUserHandlerSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
