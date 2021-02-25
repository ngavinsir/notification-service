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
