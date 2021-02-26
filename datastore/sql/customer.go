package sql

import (
	"context"
	"fmt"

	"github.com/ngavinsir/notification-service/customer"
	"gorm.io/gorm"
)

// NewCustomerRepository returns new customer respository
func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	r := &CustomerRepository{
		DB: db,
	}

	return r
}

// CustomerRepository stores customer's information
type CustomerRepository struct {
	DB *gorm.DB
}

// Save will update the customer stored in postgresql
func (r *CustomerRepository) Save(ctx context.Context, customer *customer.Customer) error {
	err := r.DB.Session(&gorm.Session{FullSaveAssociations: true}).
		Save(customer).
		Find(&customer).
		Error
	if err != nil {
		return fmt.Errorf("database error")
	}
	return nil
}

// FindByID returns customer by id
func (r *CustomerRepository) FindByID(ctx context.Context, ID uint64) (*customer.Customer, error) {
	var customer customer.Customer

	req := r.DB.
		Preload("Callback").
		Where("id = ?", ID).
		First(&customer)
	if req.Error != nil {
		return nil, fmt.Errorf("can't find customer with id: %d", ID)
	}

	return &customer, nil
}

// FindByEmail returns customer by id
func (r *CustomerRepository) FindByEmail(ctx context.Context, email string) (*customer.Customer, error) {
	var customer customer.Customer

	req := r.DB.
		Preload("Callback").
		Where("email = ?", email).
		First(&customer)
	if req.Error != nil {
		return nil, fmt.Errorf("can't find customer with email: %s", email)
	}

	return &customer, nil
}
