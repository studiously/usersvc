package usersvc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Studiously/usersvc/ddl"
	"github.com/Studiously/usersvc/templates"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"github.com/Sirupsen/logrus"
)

var (
	template = templates.NewBinTemplate(ddl.Asset, ddl.AssetDir).MustLoadDirectory("tmpl")

	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a usersvc server.
func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/signup").Handler(httptransport.NewServer(
		e.CreateUserEndpoint,
		decodeCreateUserRequest,
		encodeResponse,
		options...
	))
	r.Methods("GET").Path("/users/{id}").Handler(httptransport.NewServer(
		e.GetUserEndpoint,
		decodeGetUserRequest,
		encodeResponse,
		options...
	))
	r.Methods("GET").Path("/profiles/{userId}").Handler(httptransport.NewServer(
		e.GetProfileEndpoint,
		decodeGetProfileRequest,
		encodeResponse,
		options...
	))
	r.Methods("GET").Path("/authenticate").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challenge := r.FormValue("challenge")
		template.ExecuteTemplate(w, "authenticate.html", challenge)
	})
	return r
}

func decodeCreateUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req createUserRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		logrus.WithError(e).Info("whoops")
		return nil, e
	}
	return req, nil
}

func decodeGetUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	sid, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	id, err := uuid.FromString(sid)
	return getUserRequest{Id: id}, err
}

func decodeGetProfileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	sid, ok := vars["userId"]
	if !ok {
		return nil, ErrBadRouting
	}
	id, err := uuid.FromString(sid)
	return getProfileRequest{UserId: id}, err
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	default:
		return http.StatusInternalServerError
	}
}
