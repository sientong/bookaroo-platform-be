package routes

import (
	"github.com/bookaroo/bookaroo-platform-be/handlers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	propertyHandler := handlers.NewPropertyHandler(db)
	bookingHandler := handlers.NewBookingHandler(db)
	userHandler := handlers.NewUserHandler(db)

	// API routes
	api := r.Group("/api")
	{
		// Property routes
		properties := api.Group("/properties")
		{
			properties.GET("", propertyHandler.ListProperties)
			properties.GET("/:id", propertyHandler.GetProperty)
			properties.GET("/search", propertyHandler.SearchProperties)
			properties.POST("", propertyHandler.CreateProperty)
		}

		// Booking routes
		bookings := api.Group("/bookings")
		{
			bookings.POST("", bookingHandler.CreateBooking)
		}

		// User dashboard
		api.GET("/dashboard", userHandler.GetUserDashboard)
	}
}
