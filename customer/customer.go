package customer

// Customer stores customer's information
type Customer struct {
	BaseModel
	Email    string    `json:"email" gorm:"uniqueIndex"`
	Password string    `json:"-"`
	Callback *Callback `json:"-"`
}

// New returns new customer
func New(name, hashedPassword string) *Customer {
	return &Customer{
		Email:    name,
		Password: hashedPassword,
	}
}
