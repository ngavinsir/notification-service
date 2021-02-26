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

	if customer.ID == 0 {
		customer.ID = uint64(len(m.customerByEmail) + 1)
	}

	m.customerByEmail[customer.Email] = customer
	m.customerByID[customer.ID] = customer
	return nil
}

func (m *MockCustomerRepository) FindByID(_ context.Context, ID uint64) (*customer.Customer, error) {
	customer, ok := m.customerByID[ID]
	if !ok {
		return nil, fmt.Errorf("can't find customer with id: %d", ID)
	}
	return customer, nil
}

func (m *MockCustomerRepository) FindByEmail(_ context.Context, email string) (*customer.Customer, error) {
	customer, ok := m.customerByEmail[email]
	if !ok {
		return nil, fmt.Errorf("can't find customer with email: %s", email)
	}
	return customer, nil
}

func TestServer_Register(t *testing.T) {
	server := setupMockServer()
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
	server := setupMockServer()

	// Register customer
	if err := register(server.RegisterHandler(), "example@example.com", "password"); err != nil {
		t.Fatal(err)
	}

	loginHandler := server.LoginHandler()

	t.Run("Valid email and password", func(t *testing.T) {
		loginRequest := &AuthRequest{
			Email:    "example@example.com",
			Password: "password",
		}
		reqBody, err := json.Marshal(loginRequest)
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
		reqBody, err := json.Marshal(loginRequest)
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

func TestServer_SetCallbackURL(t *testing.T) {
	server := setupMockServer()

	// Register customer
	if err := register(server.RegisterHandler(), "example@example.com", "password"); err != nil {
		t.Fatal(err)
	}

	// Login customer
	loginResponse, err := login(server.LoginHandler(), "example@example.com", "password")
	if err != nil {
		t.Fatal(err)
	}

	handler := server.SetCallbackURLHandler()

	request := &SetCallbackURLRequest{
		CallbackURL: "http://www.example.com",
	}
	reqBody, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Response Error Unauthorized", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/callback_url", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		server.Jeff.WrapFunc(handler).ServeHTTP(rr, req)
		if statusCode := rr.Result().StatusCode; statusCode == http.StatusOK {
			t.Errorf("handler returned status code OK")
		}
	})

	t.Run("Response OK", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/callback_url", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		for i := range loginResponse.Cookies() {
			req.AddCookie(loginResponse.Cookies()[i])
		}
		server.Jeff.WrapFunc(handler).ServeHTTP(rr, req)
		if statusCode := rr.Result().StatusCode; statusCode != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", statusCode, http.StatusOK)
		}
	})

	t.Run("CallbackURL is matching", func(t *testing.T) {
		customer, err := server.CustomerRepository.FindByEmail(context.Background(), "example@example.com")
		if err != nil {
			t.Fatal(err)
		}

		if got, want := customer.Callback.CallbackURL, request.CallbackURL; got != want {
			t.Errorf("Want customer callback's url %s, got %s", want, got)
		}
	})
}

func setupMockServer() *Server {
	mockCustomerRepository := &MockCustomerRepository{
		customerByEmail: make(map[string]*customer.Customer),
		customerByID:    make(map[uint64]*customer.Customer),
	}
	jeff := jeff.New(
		memory.New(),
		jeff.Insecure,
	)
	return &Server{
		CustomerRepository: mockCustomerRepository,
		Jeff:               jeff,
	}
}

func register(handler http.HandlerFunc, email, password string) error {
	registerRequest := &AuthRequest{
		Email:    email,
		Password: password,
	}
	reqBody, err := json.Marshal(registerRequest)
	if err != nil {
		return err
	}
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	handler.ServeHTTP(rr, req)
	if statusCode := rr.Result().StatusCode; statusCode != http.StatusOK {
		return fmt.Errorf("error register")
	}

	return nil
}

func login(handler http.HandlerFunc, email, password string) (*http.Response, error) {
	loginRequest := &AuthRequest{
		Email:    email,
		Password: password,
	}
	reqBody, err := json.Marshal(loginRequest)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	handler.ServeHTTP(rr, req)
	if statusCode := rr.Result().StatusCode; statusCode != http.StatusOK {
		return nil, fmt.Errorf("error login")
	}

	return rr.Result(), nil
}
