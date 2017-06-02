package usersvc

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/ory/common/env"
	"github.com/ory/hydra/sdk"
	"github.com/studiously/introspector"
	"github.com/studiously/svcerror"
	"github.com/studiously/usersvc/codes"
	"github.com/studiously/usersvc/ddl"
	"github.com/studiously/usersvc/templates"
)

var (
	tmpls     = templates.NewBinTemplate(ddl.Asset, ddl.AssetDir).MustLoadDirectory("tmpl")
	store     = sessions.NewCookieStore([]byte(env.Getenv("COOKIE_SECRET", string(securecookie.GenerateRandomKey(32)))))
	secure, _ = strconv.ParseBool(env.Getenv("SECURE_CSRF", "false"))
	CSRF      = csrf.Protect([]byte("aNdRgUkXp2r5u8x/A?D(G+KbPeShVmYq"), csrf.Secure(secure))
	//ErrBadRouting    = errors.New("Inconsistent mapping between route and handler (programmer error).")
	//ErrPersistCookie = errors.New("Failed to add a cookie. Make sure to enable cookies.")
	//ErrInternal      = errors.New("We're having a problem on our end. Hang tight and we'll get it fixed.")
	//ErrNoChallenge   = errors.New("no challenge present")
	//ErrBadRequest    = errors.New("There was an issue parsing the request.")
	//ErrBadToken      = errors.New("Could not exchange token.")
	//ErrNotVerified   = errors.New("The consent challenge could not be verified")
	ErrInternal   = svcerror.New(codes.Nil, "internal server error")
	ErrBadRouting = svcerror.New(codes.BadRouting, "inconsistent mapping between route and handler (programmer error)")
	ErrBadRequest = svcerror.New(codes.BadRequest, "request is malformed or invalid")
)

const (
	sessionName = "authentication"
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a usersvc server.
func MakeHTTPHandler(s Service, client *sdk.Client, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(introspector.ToHTTPContext()),
	}

	r.Methods("GET").Path("/userinfo").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "users.get")(e.GetUserInfoEndpoint),
		DecodeGetUserInfoRequest,
		encodeResponse,
		options...
	))
	r.Methods("GET").Path("/users/{userID}").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "users.get")(e.GetProfileEndpoint),
		DecodeGetProfileRequest,
		encodeResponse,
		options...
	))

	r.Methods("GET").Path("/register").Handler(MakeGetRegister())
	r.Methods("POST").Path("/register").Handler(MakePostRegister(s, logger))

	r.Methods("GET").Path("/login").Handler(MakeGetLogin())
	r.Methods("POST").Path("/login").Handler(MakePostLogin(s, logger))

	r.Methods("GET").Path("/consent").Handler(MakeGetConsent(client, logger))
	r.Methods("POST").Path("/consent").Handler(MakePostConsent(client, logger))

	r.Methods("GET").Path("/logout").Handler(MakeGetLogout())

	return r
}

func MakeGetRegister() http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			tmpls.ExecuteTemplate(w, "register.html", map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(r),
				"challenge":      r.URL.Query().Get("challenge"),
				"error":          r.URL.Query().Get("error"),
			})
		}))
}

func MakePostRegister(s Service, logger log.Logger) http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				logger.Log("msg", "failed to parse form", "error", err)
				tmpls.ExecuteTemplate(w, "error.html", nil)
				return
			}
			err = s.CreateUser(r.FormValue("name"), r.FormValue("email"), r.FormValue("password"))
			if err != nil {
				logger.Log("msg", "failed to create user", "error", err)
				tmpls.ExecuteTemplate(w, "register.html", map[string]interface{}{
					csrf.TemplateTag: csrf.TemplateField(r),
					"challenge":      r.URL.Query().Get("challenge"),
					"error":          err.Error(),
				})
				return
			}
			http.Redirect(w, r, "/login?email="+r.FormValue("email")+"&challenge="+r.URL.Query().Get("challenge"), http.StatusFound)
		}))
}

func MakeGetConsent(client *sdk.Client, logger log.Logger) http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// First, check if hydra returned an error.
			err := r.URL.Query().Get("error")
			if err != "" {
				// Uh oh! There be a problem... We can just render the error template here.
				logger.Log(
					"msg", "hydra error",
					"error", r.URL.Query().Get("error"),
					"error_description", r.URL.Query().Get("error_description"))
				tmpls.ExecuteTemplate(w, "error.html", nil)
				return
			}
			// Get the challenge from the URL.
			challenge := r.URL.Query().Get("challenge")
			// Check that the challenge exists.
			if challenge == "" {
				logger.Log("msg", "consent endpoint accessed without a challenge")
				tmpls.ExecuteTemplate(w, "error.html", nil)
				return
			}
			// Check that the challenge is OK.
			claims, err2 := client.Consent.VerifyChallenge(challenge)
			if err2 != nil {
				logger.Log("msg", "challenge could not be verified", "error", err2)
				tmpls.ExecuteTemplate(w, "error.html", nil)
				return
			}
			// Check if the user is authenticated.
			user := authenticated(r)
			if user == nil {
				// Nope, not authenticated. Redirect the user to the authenticate endpoint.
				http.Redirect(w, r, "/login?challenge="+challenge, http.StatusFound)
				return
			}
			// Determine if nonconsentual is a requested scope.
			if sort.SearchStrings(claims.RequestedScopes, "nonconsentual") < len(claims.RequestedScopes) {
				// Since consent is nonconsentual, we can bypass the consent page and automatically grant all requested scopes. This will only apply to official clients.
				redirectUrl, err := client.Consent.GenerateResponse(&sdk.ResponseRequest{
					Challenge: challenge,
					// The subject is a string, usually the user id.
					Subject: user.String(),
					// The scopes our user granted.
					Scopes: claims.RequestedScopes,
				})
				// If there's a problem, we need to abort and render the error page.
				if err != nil {
					logger.Log("msg", "cannot generate response to challenge", "error", err)
					tmpls.ExecuteTemplate(w, "error.html", nil)
					return
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

func MakePostConsent(client *sdk.Client, logger log.Logger) http.Handler {
	return CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challenge := r.URL.Query().Get("challenge")
		if challenge == "" {
			logger.Log("msg", "consent endpoint called without challenge")
			tmpls.ExecuteTemplate(w, "error.html", nil)
			return
		}

		user := authenticated(r)
		if user == nil {
			http.Redirect(w, r, "/login?challenge="+challenge, http.StatusFound)
			return
		}

		if err := r.ParseForm(); err != nil {
			logger.Log("msg", "cannot parse form", "error", err)
			tmpls.ExecuteTemplate(w, "error.html", nil)
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
			logger.Log("msg", "cannot generate response to challenge", "error", err)
			tmpls.ExecuteTemplate(w, "error.html", nil)
			return
		}

		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}))
}

func MakeGetLogin() http.Handler {
	return CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challenge := r.FormValue("challenge")
		// If the user is already authenticated, redirect to the consent phase.
		session, _ := store.Get(r, sessionName)
		_, ok := session.Values["user"]
		if ok {
			http.Redirect(w, r, "/consent?challenge="+challenge, http.StatusFound)
			return
		}

		// If there is a challenge, we pass it on.
		tmpls.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"challenge":      challenge,
			"email":          r.URL.Query().Get("email"),
			csrf.TemplateTag: csrf.TemplateField(r),
		})

	}))
}

func MakePostLogin(s Service, logger log.Logger) http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				logger.Log("msg", "cannot parse form", "error", err)
				tmpls.ExecuteTemplate(w, "error.html", nil)
				return
			}
			user, err := s.Authenticate(
				r.FormValue("email"),
				r.FormValue("password"),
			)
			if err != nil {
				_, ok := err.(svcerror.Error);
				if !ok {
					logger.Log("msg", "cannot authenticate user", "error", err)
					tmpls.ExecuteTemplate(w, "error.html", nil)
					return
				}
				tmpls.ExecuteTemplate(w, "login.html", map[string]interface{}{
					"error":          err.Error(),
					"challenge":      r.URL.Query().Get("challenge"),
					csrf.TemplateTag: csrf.TemplateField(r),
				})
				return
			}
			session, _ := store.Get(r, sessionName)
			session.Values["user"] = user.String()
			if err := store.Save(r, w, session); err != nil {
				logger.Log("msg", "cannot persist session", "error", err)
				tmpls.ExecuteTemplate(w, "error.html", nil)
				return
			}
			http.Redirect(w, r, "/consent?challenge="+r.FormValue("challenge"), http.StatusFound)
		},
	))
}

func MakeGetLogout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, sessionName)
		delete(session.Values, "user")

		session.Save(r, w)

		if r.URL.Query().Get("redirect") == "" {
			tmpls.ExecuteTemplate(w, "logout.html", nil)
		} else {
			http.Redirect(w, r, r.URL.Query().Get("redirect"), http.StatusFound)
		}
	})
}

func DecodeGetUserInfoRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	return nil, nil
}

func DecodeGetProfileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	sid, ok := vars["userID"]
	if !ok {
		return nil, ErrBadRouting
	}
	id, err := uuid.Parse(sid)
	if err != nil {
		return nil, ErrInternal
	}
	request = getProfileRequest{UserID: id}
	return
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
		panic("EncodeError with nil error")
	}
	if err, ok := err.(svcerror.Error); !ok {
		err = svcerror.Wrap(codes.Nil, err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatusFrom(err.(svcerror.Error).Status()))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error_code": err.(svcerror.Error).Status(),
		"error":      err.Error(),
	})
}

func httpStatusFrom(code int) int {
	switch code {
	case codes.Nil:
		return http.StatusInternalServerError
	case codes.BadRequest:
		return http.StatusBadRequest
	case codes.WrongPassword:
		return http.StatusBadRequest
	case codes.WrongEmail:
		return http.StatusBadRequest
	case codes.HashFailed:
		return http.StatusInternalServerError
	case codes.UserExists:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func authenticated(r *http.Request) *uuid.UUID {
	session, _ := store.Get(r, sessionName)
	u, ok := session.Values["user"].(string)
	if !ok {
		return nil
	}
	user, err := uuid.Parse(u)
	if err != nil {
		return nil
	}
	return &user
}

//func rand_str(str_size int) string {
//	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
//	var bytes = make([]byte, str_size)
//	rand.Read(bytes)
//	for i, b := range bytes {
//		bytes[i] = alphanum[b%byte(len(alphanum))]
//	}
//	return string(bytes)
//}
