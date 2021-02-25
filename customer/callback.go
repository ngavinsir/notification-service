package customer

// Callback stores customer's notification callback settings
type Callback struct {
	BaseModel
	CustomerID  uint   `json:"-" gorm:"index"`
	CallbackURL string `json:"callback_url"`
}

// NewCallback returns new customer's callback settings
func NewCallback(callbackURL string, customerID uint) *Callback {
	return &Callback{
		CustomerID:  customerID,
		CallbackURL: callbackURL,
	}
}
