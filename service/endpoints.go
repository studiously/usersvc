package service

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
)

type Endpoints struct {
	GetUserEndpoint     endpoint.Endpoint
	GetProfileEndpoint  endpoint.Endpoint
	UpdateUserEndpoint  endpoint.Endpoint
	//CreateUserEndpoint  endpoint.Endpoint
	SetPasswordEndpoint endpoint.Endpoint
	//AuthenticateEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		GetUserEndpoint:     MakeGetUserEndpoint(s),
		GetProfileEndpoint:  MakeGetProfileEndpoint(s),
		UpdateUserEndpoint:  MakeUpdateUserEndpoint(s),
		//CreateUserEndpoint:  MakeCreateUserEndpoint(s),
		SetPasswordEndpoint: MakeSetPasswordEndpoint(s),
		//AuthenticateEndpoint: MakeAuthenticateEndpoint(s),
	}
}

func MakeGetUserEndpoint(s Service) endpoint.Endpoint {
	return func(c context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getUserRequest)
		user, e := s.GetUser(req.Id)
		return getUserResponse{e, user}, nil
	}
}

type getUserRequest struct {
	Id uuid.UUID `json:"id"`
}

type getUserResponse struct {
	Err  error `json:"error,omitempty"`
	User User `json:"user"`
}

func (r getUserResponse) error() error {
	return r.Err
}

func MakeGetProfileEndpoint(s Service) endpoint.Endpoint {
	return func(c context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getProfileRequest)
		profile, e := s.GetProfile(c, req.UserId)
		return getProfileResponse{e, profile}, nil
	}
}

type getProfileRequest struct {
	UserId uuid.UUID `json:"id"`
}

type getProfileResponse struct {
	Err     error `json:"error,omitempty"`
	Profile Profile `json:"user"`
}

func (r getProfileResponse) error() error {
	return r.Err
}

func MakeUpdateUserEndpoint(s Service) endpoint.Endpoint {
	return func(c context.Context, request interface{}) (response interface{}, err error) {
		req := request.(updateUserRequest)
		e := s.UpdateUser(c, req.User)
		return updateUserResponse{e}, nil
	}
}

type updateUserRequest struct {
	User User `json:"user"`
}

type updateUserResponse struct {
	Err error `json:"error"`
}

func (r updateUserResponse) error() error {
	return r.Err
}

//func MakeCreateUserEndpoint(s Service) endpoint.Endpoint {
//	return func(c context.Context, request interface{}) (response interface{}, err error) {
//		req := request.(createUserRequest)
//		user := User{
//			ID:    uuid.New(),
//			Name:  req.Name,
//			Email: req.Email,
//		}
//		e := s.CreateUser(c, user)
//		if e != nil {
//			return createUserResponse{
//				e,
//				user.ID,
//			}, nil
//		}
//		e = s.SetPassword(c, user.ID, req.Password)
//		return createUserResponse{
//			e,
//			user.ID,
//		}, nil
//	}
//}
//
//type createUserRequest struct {
//	Name     string `json:"name"`
//	Email    string `json:"email"`
//	Password string `json:"password"`
//}
//
//type createUserResponse struct {
//	Err error `json:"error,omitempty"`
//	Id  uuid.UUID `json:"id,omitempty"`
//}
//
//func (r createUserResponse) error() error {
//	return r.Err
//}

//func MakeSetUserActiveEndpoint(s Service) endpoint.Endpoint {
//	return func(c context.Context, request interface{}) (response interface{}, err error) {
//		req := request.(setUserActiveRequest)
//		e := s.DeactivateUser(c, req.UserId, req.Active)
//		return setUserActiveResponse{e}, nil
//	}
//}
//
//type setUserActiveRequest struct {
//	UserId uuid.UUID `json:"user_id"`
//	Active bool `json:"active"`
//}
//
//type setUserActiveResponse struct {
//	Err error `json:"error,omitempty"`
//}
//
//func (r setUserActiveResponse) error() error {
//	return r.Err
//}

func MakeSetPasswordEndpoint(s Service) endpoint.Endpoint {
	return func(c context.Context, request interface{}) (response interface{}, err error) {
		req := request.(setPasswordRequest)
		e := s.SetPassword(req.UserId, req.Password)
		return setPasswordResponse{e}, nil
	}
}

type setPasswordRequest struct {
	UserId   uuid.UUID
	Password string
}

type setPasswordResponse struct {
	Err error `json:"error,omitempty"`
}

func (r setPasswordResponse) error() error {
	return r.Err
}

//func MakeAuthenticateEndpoint(s Service) endpoint.Endpoint {
//	return func(c context.Context, request interface{}) (response interface{}, err error) {
//		req := request.(authenticateRequest)
//		user, e := s.Authenticate(c, req.Email, req.Password)
//		return authenticateResponse{e, user}, nil
//	}
//}
//
//type authenticateRequest struct {
//	Email    string `json:"email"`
//	Password string `json:"password"`
//}
//
//type authenticateResponse struct {
//	Err  error `json:"error"`
//	User User `json:"user"`
//}
//
//func (r authenticateResponse) error() error {
//	return r.Err
//}

//// cusr gets the user from the context
//func cusr(c context.Context) (uuid.UUID, error) {
//	claims := c.Value(jwt.JWTClaimsContextKey).(stdjwt.StandardClaims)
//	if err := claims.Valid(); err != nil {
//		return nil, err
//	}
//	id, err := strconv.Atoi(claims.Subject)
//	if err != nil {
//		return nil, ErrUnauthenticated
//	}
//	return uuid.UUID(id), nil
//}
