package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/abraithwaite/jeff"
	redis_store "github.com/abraithwaite/jeff/redis"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/gomodule/redigo/redis"
	"github.com/ngavinsir/notification-service/customer"
	"github.com/ngavinsir/notification-service/datastore"
	dssql "github.com/ngavinsir/notification-service/datastore/sql"
	"github.com/ngavinsir/notification-service/util/password"
	"gorm.io/gorm"
)

// Server holds server's required resources
type Server struct {
	CustomerRepository datastore.CustomerRepository
	Jeff               *jeff.Jeff
}

// NewServer returns new server
func NewServer(db *gorm.DB) *Server {
	redisPool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", os.Getenv("REDIS_URL")) },
	}
	sessionStore := redis_store.New(redisPool)

	return &Server{
		CustomerRepository: dssql.NewCustomerRepository(db),
		Jeff: jeff.New(
			sessionStore,
			jeff.Redirect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				render.Render(w, r, ErrUnauthorized(fmt.Errorf("invalid session")))
			})),
			jeff.Insecure,
		),
	}
}

// Router returns server routes
func (s *Server) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Post("/register", s.RegisterHandler())
	r.Post("/login", s.LoginHandler())
	r.Post("/alfamart_payment_callback", s.AlfamartPaymentCallbackHandler())

	r.Post("/callback_url", s.Jeff.WrapFunc(s.SetCallbackURLHandler()))

	return r
}

// RegisterHandler handles request for creating a customer
func (s *Server) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			render.Render(w, r, ErrBadRequest(err))
			return
		}

		hashedPassword, err := password.HashPassword(req.Password)
		if err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}

		newCustomer := customer.New(req.Email, hashedPassword)
		callback := customer.NewCallback("", uint(newCustomer.ID))
		newCustomer.Callback = callback
		if err := s.CustomerRepository.Save(r.Context(), newCustomer); err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}

		render.JSON(w, r, newCustomer)
	}
}

// LoginHandler handles request for login authentication
func (s *Server) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			render.Render(w, r, ErrBadRequest(err))
			return
		}

		customerByEmail, err := s.CustomerRepository.FindByEmail(context.Background(), req.Email)
		if err != nil {
			render.Render(w, r, ErrUnauthorized(fmt.Errorf("email/password is wrong")))
			return
		}

		if !password.CheckPasswordHash(req.Password, customerByEmail.Password) {
			render.Render(w, r, ErrUnauthorized(fmt.Errorf("email/password is wrong")))
			return
		}

		if err = s.Jeff.Set(r.Context(), w, []byte(req.Email)); err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}

		render.JSON(w, r, customerByEmail)
	}
}

// SetCallbackURLHandler handles request for setting customer's callback url
func (s *Server) SetCallbackURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SetCallbackURLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			render.Render(w, r, ErrBadRequest(err))
			return
		}

		sess := jeff.ActiveSession(r.Context())
		selectedCustomer, err := s.CustomerRepository.FindByEmail(r.Context(), string(sess.Key))
		if err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}

		selectedCustomer.Callback.CallbackURL = req.CallbackURL
		if err := s.CustomerRepository.Save(r.Context(), selectedCustomer); err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}

		render.JSON(w, r, req)
	}
}

// AlfamartPaymentCallbackHandler handles payment callback from alfamart service
func (s *Server) AlfamartPaymentCallbackHandler() http.HandlerFunc {
	type AlfamartPaymentCallbackRequest struct {
		PaymentID   string    `json:"payment_id"`
		PaymentCode string    `json:"payment_code"`
		PaidAt      time.Time `json:"paid_at"`
		ExternalID  string    `json:"external_id"`
		CustomerID  uint64    `json:"customer_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req AlfamartPaymentCallbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			render.Render(w, r, ErrBadRequest(err))
			return
		}

		involvedCustomer, err := s.CustomerRepository.FindByID(r.Context(), req.CustomerID)
		if err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}

		go GetNotifier().Notify(context.Background(), involvedCustomer, req)
		render.JSON(w, r, req)
	}
}

// AuthRequest is a struct for register and login endpoint's request body
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SetCallbackURLRequest is a struct for set callback url endpoint's request body
type SetCallbackURLRequest struct {
	CallbackURL string `json:"callback_url"`
}
