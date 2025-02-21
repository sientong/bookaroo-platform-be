package routes

import (
	_ "github.com/bookaroo/bookaroo-platform-be/docs"
	"github.com/bookaroo/bookaroo-platform-be/handlers"
	"github.com/bookaroo/bookaroo-platform-be/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
			properties.POST("", middleware.AuthMiddleware(), propertyHandler.CreateProperty)
		}

		// Booking routes
		bookings := api.Group("/bookings")
		{
			bookings.POST("", middleware.AuthMiddleware(), bookingHandler.CreateBooking)
		}

		// User routes
		api.POST("/register/owner", userHandler.RegisterOwner)
		api.POST("/register/guest", userHandler.RegisterGuest)
		api.POST("/login", userHandler.Login)

		// User dashboard
		api.GET("/dashboard", userHandler.GetUserDashboard)
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

}
