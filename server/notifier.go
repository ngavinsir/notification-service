package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ngavinsir/notification-service/customer"
)

// Notifier is an abstraction of HTTP Client that will notifies callback url by firing
// POST HTTP request
type Notifier interface {
	Notify(ctx context.Context, customer *customer.Customer, body interface{})
}

// NotifierImplementation is the default implementation of Notifier
type NotifierImplementation struct {
	HTTPClient *http.Client
}

var notifierImplementation Notifier
var httpClient *http.Client = &http.Client{}

// GetNotifier creates new notifier or returns created notifier
func GetNotifier() Notifier {
	if notifierImplementation != nil {
		return notifierImplementation
	}

	notifierImplementation = &NotifierImplementation{
		HTTPClient: httpClient,
	}

	return notifierImplementation
}

// Notify notifies customer's callback url
func (n *NotifierImplementation) Notify(
	ctx context.Context,
	customer *customer.Customer,
	body interface{},
) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("error when notifies customer %d, error: %v", customer.ID, err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		customer.Callback.CallbackURL,
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		log.Printf("error when notifies customer %d, error: %v", customer.ID, err)
	}

	resp, err := n.HTTPClient.Do(req)
	if err != nil {
		log.Printf("error when notifies customer %d, error: %v", customer.ID, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("error when notifies customer %d, error code: %v", customer.ID, resp.StatusCode)
	}
}
