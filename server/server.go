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
		customer := customer.New(req.Name)
		if err := s.customerRepository.Save(r.Context(), customer); err != nil {
			render.Render(w, r, ErrInternalServer(err))
			return
		}
		
		render.JSON(w, r, customer)
	}
}
