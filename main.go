package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ngavinsir/notification-service/customer"
	"github.com/ngavinsir/notification-service/server"
	"github.com/ngavinsir/notification-service/util/sql"
)

func main() {
	db := sql.NewGorm()
	db.AutoMigrate(
		&customer.Customer{},
		&customer.Callback{},
	)

	server := server.NewServer(db)

	port := ":4040"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	log.Printf("Server started on %s", port)
	log.Fatal(http.ListenAndServe(port, server.Router()))
}
