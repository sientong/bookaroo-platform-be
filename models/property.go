package models

// Property represents a property in the system
// @Description Property model
type Property struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Location    string          `json:"location"`
	Price       float64         `json:"price"`
	Images      []PropertyImage `json:"images" gorm:"foreignKey:PropertyID"` // Associated images
	Amenities   string          `json:"amenities"`
	OwnerID     uint            `json:"owner_id" gorm:"index"` // Foreign key for the owner
	Owner       User            `gorm:"foreignKey:OwnerID"`
	Bookings    []Booking
}

// PropertyImage represents an image associated with a property
type PropertyImage struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PropertyID uint   `json:"property_id" gorm:"index"` // Foreign key for the property
	ImageURL   string `json:"image_url"`
}
