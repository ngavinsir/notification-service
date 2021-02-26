package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abraithwaite/jeff"
	"github.com/abraithwaite/jeff/memory"
	"github.com/ngavinsir/notification-service/customer"
	. "github.com/ngavinsir/notification-service/server"
)

type MockCustomerRepository struct {
	customerByEmail map[string]*customer.Customer
	customerByID    map[uint64]*customer.Customer
}

func (m *MockCustomerRepository) Save(ctx context.Context, customer *customer.Customer) error {
	customerWithSameEmail, err := m.FindByEmail(ctx, customer.Email)
	if err == nil && customerWithSameEmail != nil && customerWithSameEmail.ID != customer.ID {
		return fmt.Errorf("violate email unique constraint")
	}

	customer.ID = uint64(len(m.customerByEmail) + 1)

	m.customerByEmail[customer.Email] = customer
	m.customerByID[customer.ID] = customer
	return nil
}

func (m *MockCustomerRepository) FindByID(_ context.Context, ID uint64) (*customer.Customer, error) {
	customer, ok := m.customerByID[ID]
	if !ok {
		return nil, nil
	}
	return customer, nil
}

func (m *MockCustomerRepository) FindByEmail(_ context.Context, email string) (*customer.Customer, error) {
	customer, ok := m.customerByEmail[email]
	if !ok {
		return nil, nil
	}
	return customer, nil
}

func TestServer_Register(t *testing.T) {
	mockCustomerRepository := &MockCustomerRepository{
		customerByEmail: make(map[string]*customer.Customer),
		customerByID:    make(map[uint64]*customer.Customer),
	}
	jeff := jeff.New(
		memory.New(),
		jeff.Insecure,
	)
	server := &Server{
		CustomerRepository: mockCustomerRepository,
		Jeff:               jeff,
	}

	handler := server.RegisterHandler()

	registerRequest := &AuthRequest{
		Email:    "example@example.com",
		Password: "password",
	}
	reqBody, err := json.Marshal(registerRequest)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Valid email", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		handler.ServeHTTP(rr, req)
		if statusCode := rr.Result().StatusCode; statusCode != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", statusCode, http.StatusOK)
		}
	})

	t.Run("Duplicate email", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		handler.ServeHTTP(rr, req)
		if statusCode := rr.Result().StatusCode; statusCode == http.StatusOK {
			t.Errorf("handler should returned error status code: got %v", statusCode)
		}
	})
}

func TestServer_Login(t *testing.T) {
	mockCustomerRepository := &MockCustomerRepository{
		customerByEmail: make(map[string]*customer.Customer),
		customerByID:    make(map[uint64]*customer.Customer),
	}
	jeff := jeff.New(
		memory.New(),
		jeff.Insecure,
	)
	server := &Server{
		CustomerRepository: mockCustomerRepository,
		Jeff:               jeff,
	}

	// Register customer
	registerHandler := server.RegisterHandler()
	registerRequest := &AuthRequest{
		Email:    "example@example.com",
		Password: "password",
	}
	reqBody, err := json.Marshal(registerRequest)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	registerHandler.ServeHTTP(rr, req)

	// Login customer
	loginHandler := server.LoginHandler()

	t.Run("Valid email and password", func(t *testing.T) {
		loginRequest := &AuthRequest{
			Email:    "example@example.com",
			Password: "password",
		}
		reqBody, err = json.Marshal(loginRequest)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		loginHandler.ServeHTTP(rr, req)
		if statusCode := rr.Result().StatusCode; statusCode != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", statusCode, http.StatusOK)
		}
	})

	t.Run("Wrong email or password", func(t *testing.T) {
		loginRequest := &AuthRequest{
			Email:    "example@example.com",
			Password: "wrong_password",
		}
		reqBody, err = json.Marshal(loginRequest)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		loginHandler.ServeHTTP(rr, req)
		if statusCode := rr.Result().StatusCode; statusCode == http.StatusOK {
			t.Errorf("handler should returned error status code: got %v, want %v", statusCode, http.StatusUnauthorized)
		}
	})
}
