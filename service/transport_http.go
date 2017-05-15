package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"sort"

	"github.com/Studiously/usersvc/ddl"
	"github.com/Studiously/usersvc/templates"
	lg"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/ory/hydra/sdk"
)

var (
	tmpls = templates.NewBinTemplate(ddl.Asset, ddl.AssetDir).MustLoadDirectory("tmpl")
	store = sessions.NewCookieStore([]byte("something-very-secret-keep-it-safe"))

	ErrBadRouting    = errors.New("inconsistent mapping between route and handler (programmer error)")
	ErrPersistCookie = errors.New("Failed to add a cookie. Make sure to enable cookies.")
	ErrNoChallenge   = errors.New("Endpoint was called without a consent challenge")
	ErrNotVerified   = errors.New("The consent challenge could not be verified")
)

const (
	sessionName = "authentication"
)

var (
	CSRF = csrf.Protect([]byte("aNdRgUkXp2r5u8x/A?D(G+KbPeShVmYq"), csrf.Secure(false))
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a usersvc server.
func MakeHTTPHandler(s Service, client *sdk.Client, logger lg.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s, client)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/users").Handler(CSRF(httptransport.NewServer(
		e.CreateUserEndpoint,
		decodeCreateUserRequest,
		encodeResponse,
		options...
	)))

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

	r.Methods("GET").Path("/me").Handler(MakeGetMe(s, client))

	r.Methods("GET").Path("/authenticate").Handler(MakeGetAuthenticate(client))
	r.Methods("POST").Path("/authenticate").Handler(MakePostAuthenticate(s, client))

	r.Methods("GET").Path("/consent").Handler(MakeGetConsent(client))
	r.Methods("POST").Path("/consent").Handler(MakePostConsent(client))

	r.Methods("GET").Path("/signout").Handler(MakeGetSignout())

	return r
}

func MakeGetConsent(client *sdk.Client) http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			error := r.URL.Query().Get("error")
			if error != "" {
				tmpls.ExecuteTemplate(w, "error.html", map[string]interface{}{
					"error":            error,
					"errorDescription": r.URL.Query().Get("error_description"),
				})
				return
			}
			challenge := r.URL.Query().Get("challenge")
			if challenge == "" {
				encodeError(nil, ErrNoChallenge, w)
				return
			}
			claims, err := client.Consent.VerifyChallenge(challenge)
			if err != nil {
				encodeError(nil, ErrNotVerified, w)
				return
			}

			user := authenticated(r)
			if user == uuid.Nil {
				http.Redirect(w, r, "/authenticate?challenge="+challenge, http.StatusFound)
				return
			}

			if sort.SearchStrings(claims.RequestedScopes, "nonconsentual") < len(claims.RequestedScopes) {
				redirectUrl, err := client.Consent.GenerateResponse(&sdk.ResponseRequest{
					Challenge: challenge,
					// The subject is a string, usually the user id.
					Subject: user.String(),
					// The scopes our user granted.
					Scopes: claims.RequestedScopes,
				})
				if err != nil {
					encodeError(nil, err, w)
				}
				http.Redirect(w, r, redirectUrl, http.StatusFound)
				return
			}

			tmpls.ExecuteTemplate(w, "consent.html", struct {
				*sdk.ChallengeClaims
				Challenge string
				csrfField template.HTML
			}{ChallengeClaims: claims, Challenge: challenge, csrfField: csrf.TemplateField(r) })
		}))
}

func MakePostConsent(client *sdk.Client) http.Handler {
	return CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challenge := r.URL.Query().Get("challenge")
		if challenge == "" {
			encodeError(nil, ErrNoChallenge, w)
			return
		}

		user := authenticated(r)
		if user == uuid.Nil {
			http.Redirect(w, r, "/authenticate?challenge="+challenge, http.StatusFound)
			return
		}

		if err := r.ParseForm(); err != nil {
			encodeError(nil, err, w)
			return
		}

		var grantedScopes = []string{}
		for key := range r.PostForm {
			// And add each scope to the list of granted scopes.
			grantedScopes = append(grantedScopes, key)
		}
		redirectUrl, err := client.Consent.GenerateResponse(&sdk.ResponseRequest{
			Challenge: challenge,

			// The subject is a string, usually the user id.
			Subject: user.String(),

			// The scopes our user granted.
			Scopes: grantedScopes,
		})
		if err != nil {
			encodeError(nil, err, w)
		}

		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}))
}

func MakeGetAuthenticate(client *sdk.Client) http.Handler {
	return CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If there is no challenge, we err out.
		challenge := r.FormValue("challenge")
		if challenge == "" {
			encodeError(nil, ErrNoChallenge, w)
			return
		}

		// If the user is already authenticated, redirect to the consent phase.
		session, _ := store.Get(r, sessionName)
		_, ok := session.Values["user"]
		if ok {
			http.Redirect(w, r, "/consent?challenge="+challenge, http.StatusFound)
			return
		}

		// If there is a challenge, we pass it on.
		tmpls.ExecuteTemplate(w, "authenticate.html", map[string]interface{}{
			"challenge":      challenge,
			csrf.TemplateTag: csrf.TemplateField(r),
		})

	}))
}

func MakePostAuthenticate(s Service, client *sdk.Client) http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				encodeError(nil, err, w)
				return
			}
			user, err := s.Authenticate(
				r.FormValue("email"),
				r.FormValue("password"),
			)
			if err != nil {
				encodeError(nil, err, w)
				return
			}
			//r := response.(authenticateResponse)
			//if r.error() != nil {
			//	encodeError(c, r.error(), w)
			//	return nil
			//}
			session, _ := store.Get(r, sessionName)
			session.Values["user"] = user.ID.String()

			if err := store.Save(r, w, session); err != nil {
				encodeError(nil, err, w)
				return
			}
			challenge := r.FormValue("challenge")
			http.Redirect(w, r, "/consent?challenge="+challenge, http.StatusFound)
		},
	))
}

func MakeGetMe(s Service, client *sdk.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authenticated(r) == uuid.Nil {
			nonce := rand_str(24)
			var authUrl = client.OAuth2Config("http://localhost:8080/me", "offline", "nonconsentual", "openid").AuthCodeURL(nonce) + "&nonce=" + nonce
			http.Redirect(w, r, authUrl, http.StatusFound)
			return
		}
		uid := authenticated(r)
		user, err := s.GetUser(uid)
		if err != nil {
			tmpls.ExecuteTemplate(w, "error.html", map[string]interface{}{
				"error":            "Internal Server Error",
				"errorDescription": err.Error(),
			})
			return
		}
		tmpls.ExecuteTemplate(w, "me.html", user)
	})
}
func MakeGetSignout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, sessionName)
		delete(session.Values, "user")
		session.Save(r, w)
	})
}

func decodeCreateUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	//var req createUserRequest
	//json.NewDecoder(r.Body).Decode(&req)
	//return req, err
	err = r.ParseForm()
	if err != nil {
		return
	}
	request = createUserRequest{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	return
}

func decodeGetUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	sid, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	id, err := uuid.Parse(sid)
	request = getUserRequest{Id: id}
	return
}

func decodeGetProfileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	sid, ok := vars["userId"]
	if !ok {
		return nil, ErrBadRouting
	}
	id, err := uuid.Parse(sid)
	request = getProfileRequest{UserId: id}
	return
}

//func decodeAuthenticateRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
//
//	return
//}

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
	case ErrUserExists:
		return http.StatusBadRequest
	case ErrNoChallenge:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func authenticated(r *http.Request) uuid.UUID {
	session, _ := store.Get(r, sessionName)
	u, ok := session.Values["user"].(string);
	if !ok {
		return uuid.Nil
	}
	user, err := uuid.Parse(u);
	if err != nil {
		return uuid.Nil
	}
	return user
}

func rand_str(str_size int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, str_size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
