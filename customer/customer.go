package customer

// Customer stores customer's information
type Customer struct {
	BaseModel
	Name     string    `json:"name"`
	Callback *Callback `json:"-"`
}

// New returns new customer
func New(name string) *Customer {
	return &Customer{
		Name: name,
	}
}
