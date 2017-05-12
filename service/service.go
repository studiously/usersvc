package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrUserExists      = errors.New("Oh noes! That email's taken.")
	ErrWrongEmail      = errors.New("Sorry, nobody on my list with that email.")
	ErrWrongPassword   = errors.New("Drats, wrong password. Maybe the other one?")
	ErrUnauthenticated = errors.New("must be authenticated")
	ErrBadEmail        = errors.New("bad email format")
	ErrHashFailed      = errors.New("bcrypt hash failed")
)

type Service interface {
	GetUser(c context.Context, id uuid.UUID) (User, error)
	GetProfile(c context.Context, id uuid.UUID) (Profile, error)
	CreateUser(c context.Context, user User) (err error)
	UpdateUser(c context.Context, user User) error
	SetPassword(c context.Context, userId uuid.UUID, password string) error
	Authenticate(email string, password string) (User, error)
}

type User struct {
	ID     uuid.UUID `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool `json:"active"`
}

type Profile struct {
	Name string `json:"name"`
}
