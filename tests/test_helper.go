package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres password=your_password dbname=bookaroo_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

// CreateTestContext creates a new Gin context for testing
func CreateTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// ParseResponse parses the JSON response body into the given interface
func ParseResponse(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), v)
	assert.NoError(t, err)
}

// MakeRequest creates and executes a test HTTP request
func MakeRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// MakeRequestWithToken makes an HTTP request with the given method, path, and body, including a JWT token in the Authorization header.
func MakeRequestWithToken(router *gin.Engine, method, path string, body []byte, token string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token) // Set the token in the Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ClearTestDB clears the test database
func ClearTestDB(t *testing.T, db *gorm.DB) {
	err := db.Exec("DELETE FROM users").Error
	if err != nil {
		t.Fatalf("failed to clear test DB: %v", err)
	}
}
