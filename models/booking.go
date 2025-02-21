package models

import (
	"time"
)

// Booking represents a booking in the system
// @Description Booking model
type Booking struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	PropertyID uint      `json:"property_id" gorm:"index"`
	Property   Property  `gorm:"foreignKey:PropertyID"`
	UserID     uint      `json:"user_id" gorm:"index"`
	User       User      `gorm:"foreignKey:UserID"`
	StartDate  time.Time `json:"start_date" gorm:"index"`
	EndDate    time.Time `json:"end_date" gorm:"index"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status" gorm:"default:'pending'"`
}
