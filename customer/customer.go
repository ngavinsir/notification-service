package customer

import "gorm.io/gorm"

// Customer stores customer's information
type Customer struct {
	gorm.Model
	Callback Callback
}

// New returns new customer
func New() *Customer {
	return &Customer{}
}
