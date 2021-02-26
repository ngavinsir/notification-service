package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

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

	email := "example@example.com"
	password := "password"

	t.Run("Valid email", func(t *testing.T) {
		err := mustRegister(handler, email, password)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Duplicate email", func(t *testing.T) {
		registerResponse, err := register(handler, email, password)
		if err != nil {
			t.Fatal(err)
		}

		if statusCode := registerResponse.StatusCode; statusCode == http.StatusOK {
			t.Errorf("handler should returned error status code: got %v", statusCode)
		}
	})
}

func TestServer_Login(t *testing.T) {
	server := setupMockServer()

	// Register customer
	if err := mustRegister(server.RegisterHandler(), "example@example.com", "password"); err != nil {
		t.Fatal(err)
	}

	loginHandler := server.LoginHandler()

	t.Run("Valid email and password", func(t *testing.T) {
		_, err := mustLogin(loginHandler, "example@example.com", "password")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Wrong email or password", func(t *testing.T) {
		loginResponse, err := login(loginHandler, "example@example.com", "wrong_password")
		if err != nil {
			t.Fatal(err)
		}

		if statusCode := loginResponse.StatusCode; statusCode == http.StatusOK {
			t.Errorf("handler should returned error status code: got %v, want %v", statusCode, http.StatusUnauthorized)
		}
	})
}

func TestServer_SetCallbackURL(t *testing.T) {
	server := setupMockServer()

	// Register customer
	if err := mustRegister(server.RegisterHandler(), "example@example.com", "password"); err != nil {
		t.Fatal(err)
	}

	// Login customer
	loginResponse, err := mustLogin(server.LoginHandler(), "example@example.com", "password")
	if err != nil {
		t.Fatal(err)
	}

	handler := server.SetCallbackURLHandler()

	t.Run("Response Error Unauthorized", func(t *testing.T) {
		setCallbackURLResponse, err := setCallbackURL(
			server.Jeff.WrapFunc(handler),
			"http://www.example.com",
			[]*http.Cookie{},
		)
		if err != nil {
			t.Fatal(err)
		}

		if statusCode := setCallbackURLResponse.StatusCode; statusCode == http.StatusOK {
			t.Errorf("handler returned status code OK")
		}
	})

	t.Run("Response OK", func(t *testing.T) {
		err := mustSetCallbackURL(
			server.Jeff.WrapFunc(handler),
			"http://www.example.com",
			loginResponse.Cookies(),
		)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CallbackURL is matching", func(t *testing.T) {
		customer, err := server.CustomerRepository.FindByEmail(context.Background(), "example@example.com")
		if err != nil {
			t.Fatal(err)
		}

		if got, want := customer.Callback.CallbackURL, "http://www.example.com"; got != want {
			t.Errorf("Want customer callback's url %s, got %s", want, got)
		}
	})
}

func TestServer_AlfamartPaymentCallback(t *testing.T) {
	var wg sync.WaitGroup
	server := setupMockServer()

	paidAt, _ := time.Parse(time.RFC3339, "2020-10-17T07:41:33.866Z")
	alfamartRequest := &AlfamartPaymentCallbackRequest{
		PaymentID:   "123123123",
		PaymentCode: "XYZ123",
		PaidAt:      paidAt,
		ExternalID:  "order-123",
		CustomerID:  1,
	}

	mockCustomerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req AlfamartPaymentCallbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		t.Run("Notification payload is matching", func(t *testing.T) {
			if got, want := req, *alfamartRequest; got != want {
				t.Errorf("Want notification payload %v, got %v", want, got)
			}
		})

		w.Write([]byte(`OK`))
		wg.Done()
	}))
	defer mockCustomerServer.Close()

	// Set callback url
	if err := mustRegister(server.RegisterHandler(), "example@example.com", "password"); err != nil {
		t.Fatal(err)
	}
	loginResponse, err := mustLogin(server.LoginHandler(), "example@example.com", "password")
	if err != nil {
		t.Fatal(err)
	}
	err = mustSetCallbackURL(
		server.Jeff.WrapFunc(server.SetCallbackURLHandler()),
		mockCustomerServer.URL,
		loginResponse.Cookies(),
	)
	if err != nil {
		t.Fatal(err)
	}

	handler := server.AlfamartPaymentCallbackHandler()

	wg.Add(1)
	_, err = sendRequest(handler, "POST", "/alfamart_payment_callback", alfamartRequest, []*http.Cookie{})
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}

func setupMockServer() *Server {
	return &Server{
		CustomerRepository: &MockCustomerRepository{
			customerByEmail: make(map[string]*customer.Customer),
			customerByID:    make(map[uint64]*customer.Customer),
		},
		Jeff: jeff.New(
			memory.New(),
			jeff.Insecure,
		),
	}
}

func mustRegister(handler http.HandlerFunc, email, password string) error {
	registerResponse, err := register(handler, email, password)
	if err != nil {
		return err
	}
	if statusCode := registerResponse.StatusCode; statusCode != http.StatusOK {
		return fmt.Errorf("error register")
	}
	return nil
}

func mustLogin(handler http.HandlerFunc, email, password string) (*http.Response, error) {
	loginResponse, err := login(handler, email, password)
	if err != nil {
		return nil, err
	}
	if statusCode := loginResponse.StatusCode; statusCode != http.StatusOK {
		return nil, fmt.Errorf("error login")
	}
	return loginResponse, nil
}

func mustSetCallbackURL(handler http.HandlerFunc, callbackURL string, cookies []*http.Cookie) error {
	response, err := setCallbackURL(handler, callbackURL, cookies)
	if err != nil {
		return err
	}
	if statusCode := response.StatusCode; statusCode != http.StatusOK {
		return fmt.Errorf("error set callback url")
	}
	return nil
}

func register(handler http.HandlerFunc, email, password string) (*http.Response, error) {
	registerRequest := &AuthRequest{
		Email:    email,
		Password: password,
	}
	return sendRequest(handler, "POST", "/register", registerRequest, []*http.Cookie{})
}

func login(handler http.HandlerFunc, email, password string) (*http.Response, error) {
	loginRequest := &AuthRequest{
		Email:    email,
		Password: password,
	}
	return sendRequest(handler, "POST", "/login", loginRequest, []*http.Cookie{})
}

func setCallbackURL(handler http.HandlerFunc, callbackURL string, cookies []*http.Cookie) (*http.Response, error) {
	request := &SetCallbackURLRequest{
		CallbackURL: callbackURL,
	}
	return sendRequest(handler, "POST", "/callback_url", request, cookies)
}

func sendRequest(
	handler http.HandlerFunc,
	method string,
	url string,
	payload interface{},
	cookies []*http.Cookie,
) (*http.Response, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	for i := range cookies {
		req.AddCookie(cookies[i])
	}
	handler.ServeHTTP(rr, req)

	return rr.Result(), nil
}
