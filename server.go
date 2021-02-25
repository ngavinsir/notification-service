package main

import (
	"github.com/ngavinsir/notification-service/customer"
	"github.com/ngavinsir/notification-service/util/sql"
)

func main() {
	db := sql.NewGorm()
	db.AutoMigrate(
		&customer.Customer{},
		&customer.Callback{},
	)
}
