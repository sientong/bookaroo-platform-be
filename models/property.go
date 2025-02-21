package models

import (
	"gorm.io/gorm"
)

type Property struct {
	gorm.Model
	Name        string
	Description string
	Location    string
	Price       float64
	Images      []PropertyImage
	Amenities   string
	OwnerID     uint
	Owner       User `gorm:"foreignKey:OwnerID"`
	Bookings    []Booking
}

type PropertyImage struct {
	gorm.Model
	PropertyID uint
	ImageURL   string
}
