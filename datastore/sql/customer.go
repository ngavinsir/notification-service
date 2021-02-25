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
	if err := r.DB.Save(customer).Find(&customer).Error; err != nil {
		return fmt.Errorf("database error")
	}
	return nil
}

// FindByID returns customer by id
func (r *CustomerRepository) FindByID(ctx context.Context, ID uint) (*customer.Customer, error) {
	var customer customer.Customer

	req := r.DB.
		Preload("Callback").
		Where("id = ?", ID).
		First(&customer)
	if req.Error != nil {
		return nil, req.Error
	}

	return &customer, nil
}
