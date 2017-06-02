package usersvc

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/go-nats"
	"github.com/studiously/usersvc/models"
)

const (
	SubjDeleteUser = "users.delete"
)

func MessagingMiddleware(nc *nats.Conn) Middleware {
	return func(next Service) Service {
		return messagingMiddleware{nc, next}
	}
}

type messagingMiddleware struct {
	nc   *nats.Conn
	next Service
}

func (mm messagingMiddleware) GetProfile(ctx context.Context, userID uuid.UUID) (name string, err error) {
	return mm.next.GetProfile(ctx, userID)
}

func (mm messagingMiddleware) GetUserInfo(ctx context.Context) (user *models.User, err error) {
	return mm.next.GetUserInfo(ctx)
}

func (mm messagingMiddleware) CreateUser(name, email, password string) error {
	return mm.next.CreateUser(name, email, password)
}

func (mm messagingMiddleware) SetName(ctx context.Context, name string) error {
	return mm.next.SetName(ctx, name)
}

func (mm messagingMiddleware) SetEmail(ctx context.Context, email string) error {
	return mm.next.SetEmail(ctx, email)
}

func (mm messagingMiddleware) SetPassword(ctx context.Context, password string) error {
	return mm.next.SetPassword(ctx, password)
}

func (mm messagingMiddleware) Authenticate(email string, password string) (uuid.UUID, error) {
	return mm.next.Authenticate(email, password)
}

func (mm messagingMiddleware) DeleteUser(ctx context.Context) (err error) {
	defer func() {
		if err == nil {
			// Don't care about errors because worst-case scenario, some data doesn't get deleted.
			// It's inaccessible and thus irrelevant.
			userID, _ := subj(ctx).MarshalText()
			mm.nc.Publish(SubjDeleteUser, userID)
		}
	}()
	return mm.next.DeleteUser(ctx)
}
