package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"

	lg"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/ory/common/env"
	"github.com/ory/hydra/sdk"
	"github.com/studiously/usersvc/ddl"
	"github.com/studiously/usersvc/templates"
)

var (
	tmpls            = templates.NewBinTemplate(ddl.Asset, ddl.AssetDir).MustLoadDirectory("tmpl")
	store            = sessions.NewCookieStore([]byte(env.Getenv("COOKIE_SECRET", string(securecookie.GenerateRandomKey(32)))))
	secure, _        = strconv.ParseBool(env.Getenv("SECURE_CSRF", "false"))
	CSRF             = csrf.Protect([]byte("aNdRgUkXp2r5u8x/A?D(G+KbPeShVmYq"), csrf.Secure(secure))
	ErrBadRouting    = errors.New("Inconsistent mapping between route and handler (programmer error).")
	ErrPersistCookie = errors.New("Failed to add a cookie. Make sure to enable cookies.")
	ErrInternal      = errors.New("We're having a problem on our end. Hang tight and we'll get it fixed.")
	ErrNoChallenge   = errors.New("Endpoint was called without a consent challenge")
	ErrBadRequest    = errors.New("There was an issue parsing the request.")
	//ErrNotVerified   = errors.New("The consent challenge could not be verified")
)

const (
	sessionName = "authentication"
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a usersvc server.
func MakeHTTPHandler(s Service, client *sdk.Client, logger lg.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	//r.Methods("POST").Path("/users").Handler(CSRF(httptransport.NewServer(
	//	e.CreateUserEndpoint,
	//	decodeCreateUserRequest,
	//	encodeResponse,
	//	options...
	//)))

	r.Methods("GET").Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/me", http.StatusPermanentRedirect)
	})

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

	r.Methods("GET").Path("/register").Handler(MakeGetRegister())
	r.Methods("POST").Path("/register").Handler(MakePostRegister(s))

	r.Methods("GET").Path("/login").Handler(MakeGetLogin(client))
	r.Methods("POST").Path("/login").Handler(MakePostLogin(s))

	r.Methods("GET").Path("/consent").Handler(MakeGetConsent(client))
	r.Methods("POST").Path("/consent").Handler(MakePostConsent(client))

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

func MakePostRegister(s Service) http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				tmpls.ExecuteTemplate(w, "register.html", map[string]interface{}{
					csrf.TemplateTag: csrf.TemplateField(r),
					"challenge":      r.URL.Query().Get("challenge"),
					"error":          ErrBadRequest.Error(),
				})
				return
			}
			user := User{
				ID:    uuid.New(),
				Name:  r.FormValue("name"),
				Email: r.FormValue("email"),
			}
			err = s.CreateUser(user)
			if err != nil {
				tmpls.ExecuteTemplate(w, "register.html", map[string]interface{}{
					csrf.TemplateTag: csrf.TemplateField(r),
					"challenge":      r.URL.Query().Get("challenge"),
					"error":          ErrInternal.Error(),
				})
				return
			}
			err = s.SetPassword(user.ID, r.FormValue("password"))
			if err != nil {
				tmpls.ExecuteTemplate(w, "register.html", map[string]interface{}{
					csrf.TemplateTag: csrf.TemplateField(r),
					"challenge":      r.URL.Query().Get("challenge"),
					"error":          ErrInternal.Error(),
				})
				return
			}
			http.Redirect(w, r, "/login?email="+user.Email+"&challenge="+r.URL.Query().Get("challenge"), http.StatusFound)
		}))
}

func MakeGetConsent(client *sdk.Client) http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// First, check if hydra returned an error.
			err := r.URL.Query().Get("error")
			if err != "" {
				// Uh oh! There be a problem here. We can just render the error template here.
				tmpls.ExecuteTemplate(w, "error.html", map[string]interface{}{
					"error":            err,
					"errorDescription": r.URL.Query().Get("error_description"),
				})
				return
			}
			// Get the challenge from the URL.
			challenge := r.URL.Query().Get("challenge")
			// Check that the challenge exists.
			if challenge == "" {
				// Without a challenge, we can't authenticate--default to redirecting the user to /me to initialize the process. This makes it work even without a challenge--it just defaults to the user profile app.
				http.Redirect(w, r, "/me", http.StatusFound)
				return
			}
			// Check that the challenge is OK.
			claims, err2 := client.Consent.VerifyChallenge(challenge)
			if err2 != nil {
				// If the challenge is not OK, redirect to /me again.
				http.Redirect(w, r, "/me", http.StatusFound)
				return
			}
			// Check if the user is authenticated.
			user := authenticated(r)
			if user == uuid.Nil {
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
					encodeError(nil, err, w)
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

func MakePostConsent(client *sdk.Client) http.Handler {
	return CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challenge := r.URL.Query().Get("challenge")
		if challenge == "" {
			encodeError(nil, ErrNoChallenge, w)
			return
		}

		user := authenticated(r)
		if user == uuid.Nil {
			http.Redirect(w, r, "/login?challenge="+challenge, http.StatusFound)
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
			return
		}

		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}))
}

func MakeGetLogin(client *sdk.Client) http.Handler {
	return CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challenge := r.FormValue("challenge")
		//if challenge == "" {
		//	http.Redirect(w, r, "/me", http.StatusFound)
		//	return
		//}
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

func MakePostLogin(s Service) http.Handler {
	return CSRF(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				tmpls.ExecuteTemplate(w, "login.html", map[string]interface{}{
					"error":          err.Error(),
					"challenge":      r.URL.Query().Get("challenge"),
					csrf.TemplateTag: csrf.TemplateField(r),

				})
				return
			}
			user, err := s.Authenticate(
				r.FormValue("email"),
				r.FormValue("password"),
			)
			if err != nil {
				tmpls.ExecuteTemplate(w, "login.html", map[string]interface{}{
					"error":          err.Error(),
					"challenge":      r.URL.Query().Get("challenge"),
					csrf.TemplateTag: csrf.TemplateField(r),
				})
				return
			}
			session, _ := store.Get(r, sessionName)
			session.Values["user"] = user.ID.String()
			if err := store.Save(r, w, session); err != nil {
				tmpls.ExecuteTemplate(w, "login.html", map[string]interface{}{
					"error":          ErrPersistCookie.Error(),
					"challenge":      r.URL.Query().Get("challenge"),
					csrf.TemplateTag: csrf.TemplateField(r),
				})
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
			var authUrl = client.OAuth2Config(os.Getenv("REDIRECT_URL"), "offline", "nonconsentual", "openid").AuthCodeURL(nonce) + "&nonce=" + nonce
			http.Redirect(w, r, authUrl, http.StatusFound)
			return
		}
		uid := authenticated(r)
		user, err := s.GetUser(uid)
		if err != nil {
			http.Redirect(w, r, "/logout", http.StatusFound)
		}
		tmpls.ExecuteTemplate(w, "me.html", map[string]interface{}{
			"user": user,
		})
	})
}
func MakeGetLogout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, sessionName)
		delete(session.Values, "user")
		session.Save(r, w)
		http.Redirect(w, r, "/me", http.StatusFound)
	})
}

//func decodeCreateUserRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
//	//var req createUserRequest
//	//json.NewDecoder(r.Body).Decode(&req)
//	//return req, err
//	err = r.ParseForm()
//	if err != nil {
//		return
//	}
//	request = createUserRequest{
//		Name:     r.FormValue("name"),
//		Email:    r.FormValue("email"),
//		Password: r.FormValue("password"),
//	}
//	return
//}

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

func renderError(w http.ResponseWriter, error, description string) {

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
	u, ok := session.Values[sessionName].(string);
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
