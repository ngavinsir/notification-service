package customer

import "gorm.io/gorm"

// Callback stores customer's notification callback settings
type Callback struct {
	gorm.Model
	CustomerID  uint   `gorm:"index"`
	CallbackURL string `json:"callback_url"`
}

// NewCallback returns new customer's callback settings
func NewCallback(callbackURL string, customerID uint) *Callback {
	return &Callback{
		CustomerID:  customerID,
		CallbackURL: callbackURL,
	}
}
