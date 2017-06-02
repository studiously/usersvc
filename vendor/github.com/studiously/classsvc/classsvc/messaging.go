package classsvc

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/go-nats"
	"github.com/studiously/classsvc/models"
)

const (
	SubjDeleteClass = "classes.delete"
	SubjLeaveClass  = "classes.leave"
)

func MessagingMiddleware(nc *nats.Conn) (Middleware, error) {
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}
	return func(next Service) Service {
		return messagingMiddleware{ec, next}
	}, nil
}

type messagingMiddleware struct {
	nc   *nats.EncodedConn
	next Service
}

func (mm messagingMiddleware) ListClasses(ctx context.Context) ([]uuid.UUID, error) {
	return mm.next.ListClasses(ctx)
}

func (mm messagingMiddleware) GetClass(ctx context.Context, classID uuid.UUID) (*models.Class, error) {
	return mm.next.GetClass(ctx, classID)
}

func (mm messagingMiddleware) CreateClass(ctx context.Context, name string) (*uuid.UUID, error) {
	return mm.next.CreateClass(ctx, name)
}

func (mm messagingMiddleware) UpdateClass(ctx context.Context, classID uuid.UUID, name *string, currentUnit *uuid.UUID) error {
	return mm.next.UpdateClass(ctx, classID, name, currentUnit)
}

func (mm messagingMiddleware) DeleteClass(ctx context.Context, classID uuid.UUID) (err error) {
	defer func() {
		if err == nil {
			mm.nc.Publish(SubjDeleteClass, struct {
				ClassID uuid.UUID `json:"class_id"`
			}{classID})
		}
	}()
	return mm.next.DeleteClass(ctx, classID)
}

func (mm messagingMiddleware) JoinClass(ctx context.Context, classID uuid.UUID) (error) {
	return mm.next.JoinClass(ctx, classID)
}

func (mm messagingMiddleware) LeaveClass(ctx context.Context, userID *uuid.UUID, classID uuid.UUID) (err error) {
	defer func() {
		if err == nil {
			mm.nc.Publish(SubjLeaveClass, struct {
				ClassID uuid.UUID `json:"class_id"`
				UserID  *uuid.UUID `json:"user_id,omitempty"`
			}{classID, userID})
		}
	}()
	return mm.next.LeaveClass(ctx, userID, classID)
}

func (mm messagingMiddleware) SetRole(ctx context.Context, classID, userID uuid.UUID, role models.UserRole) error {
	return mm.next.SetRole(ctx, classID, userID, role)
}

func (mm messagingMiddleware) ListMembers(ctx context.Context, classID uuid.UUID) ([]*models.Member, error) {
	return mm.next.ListMembers(ctx, classID)
}

func (mm messagingMiddleware) GetMember(ctx context.Context, classID, userID uuid.UUID) (member *models.Member, err error) {
	return mm.next.GetMember(ctx, classID, userID)
}
