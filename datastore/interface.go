package datastore

import (
	"context"

	"github.com/ngavinsir/notification-service/customer"
)

// CustomerRepository is an interface for customer storage
type CustomerRepository interface {
	Save(ctx context.Context, customer *customer.Customer) error
}
