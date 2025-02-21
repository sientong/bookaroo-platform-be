package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email        string  `json:"email" gorm:"unique"`
	Password     string  `json:"password"`
	Name         string  `json:"name"`
	Role         string  `json:"role"`
	Phone        string  `json:"phone"`         // Added
	Address      string  `json:"address"`       // Added
	BusinessName *string `json:"business_name"` // Added as pointer since it's optional (only for owners)
}
