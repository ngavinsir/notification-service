package server

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrResponse contains err, http_status_code, status_text, app_code, error_text
type ErrResponse struct {
	Err            error `json:"-"`
	HTTPStatusCode int   `json:"-"`

	StatusText string `json:"status"`
	ErrorText  string `json:"error,omitempty"`
}

// Render error response
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrBadRequest returns bad request error response
func ErrBadRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "bad request",
		ErrorText:      err.Error(),
	}
}

// ErrInternalServer returns internal server error response
func ErrInternalServer(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "internal error",
		ErrorText:      err.Error(),
	}
}
