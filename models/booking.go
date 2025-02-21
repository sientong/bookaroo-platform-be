package models

import (
	"time"
	"gorm.io/gorm"
)

type Booking struct {
	gorm.Model
	PropertyID  uint
	Property    Property
	UserID      uint
	User        User
	StartDate   time.Time
	EndDate     time.Time
	TotalPrice  float64
	Status      string // "pending", "confirmed", "cancelled", "completed"
}
