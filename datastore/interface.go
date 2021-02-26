package datastore

import (
	"context"

	"github.com/ngavinsir/notification-service/customer"
)

// CustomerRepository is an interface for customer storage
type CustomerRepository interface {
	Save(ctx context.Context, customer *customer.Customer) error
	FindByID(ctx context.Context, ID uint64) (*customer.Customer, error)
	FindByEmail(ctx context.Context, email string) (*customer.Customer, error)
}
