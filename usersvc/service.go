package usersvc

import (
	"context"

	"github.com/google/uuid"
	"github.com/studiously/svcerror"
	"github.com/studiously/usersvc/codes"
	"github.com/studiously/usersvc/models"
)

var (
	ErrHashFailed    = svcerror.New(codes.HashFailed, "bcrypt hash failed")
	ErrUserExists    = svcerror.New(codes.UserExists, "user already exists")
	ErrWrongEmail    = svcerror.New(codes.WrongEmail, "wrong email")
	ErrWrongPassword = svcerror.New(codes.WrongPassword, "wrong password")
	ErrNotFound      = svcerror.New(codes.NotFound, "not found")
	ErrDeleteOwner   = svcerror.New(codes.DeleteOwner, "cannot delete user while it is an owner")
)

type Middleware func(Service) Service

type Service interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (name string, err error)
	GetUserInfo(ctx context.Context) (user *models.User, err error)
	CreateUser(name, email, password string) error
	SetName(ctx context.Context, name string) error
	SetEmail(ctx context.Context, email string) error
	SetPassword(ctx context.Context, password string) error
	Authenticate(email string, password string) (uuid.UUID, error)
	DeleteUser(ctx context.Context) error
}
