package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email      string `gorm:"unique;not null"`
	Password   string `json:"-"`
	Name       string
	Role       string     // "owner" or "guest"
	Properties []Property `gorm:"foreignKey:OwnerID"`
	Bookings   []Booking
}
