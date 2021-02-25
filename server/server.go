package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/ngavinsir/notification-service/customer"
	"github.com/ngavinsir/notification-service/datastore"
	dssql "github.com/ngavinsir/notification-service/datastore/sql"
	"gorm.io/gorm"
)

// Server holds repository
type Server struct {
	customerRepository datastore.CustomerRepository
}

// NewServer returns new server
func NewServer(db *gorm.DB) *Server {
	return &Server{
		customerRepository: dssql.NewCustomerRepository(db),
	}
}

// Router returns server routes
func (s *Server) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/customer", s.CreateCustomerHandler())
	r.Post("/callback_url", s.SetCallbackURLHandler())

	return r
}

// CreateCustomerHandler handles request for creating a customer
func (s *Server) CreateCustomerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type CreateCustomerRequest struct {
			Name string `json:"name"`
		}

		var req CreateCustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			render.Render(w, r, ErrBadRequest(err))
			return
		}

		newCustomer := customer.New(req.Name)
		callback := customer.NewCallback("", uint(newCustomer.ID))
		newCustomer.Callback = callback
		if err := s.customerRepository.Save(r.Context(), newCustomer); err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}
		
		render.JSON(w, r, newCustomer)
	}
}

// SetCallbackURLHandler handles request for setting customer's callback url
func (s *Server) SetCallbackURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type SetCallbackURLRequest struct {
			CustomerID uint `json:"customer_id"`
			CallbackURL string `json:"callback_url"`
		}

		var req SetCallbackURLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			render.Render(w, r, ErrBadRequest(err))
			return
		}

		selectedCustomer, err := s.customerRepository.FindByID(r.Context(), req.CustomerID)
		if err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}

		selectedCustomer.Callback.CallbackURL = req.CallbackURL
		if err := s.customerRepository.Save(r.Context(), selectedCustomer); err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}
		
		render.JSON(w, r, req)
	}
}