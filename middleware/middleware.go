package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/studiously/introspector"
	"github.com/studiously/usersvc/usersvc"
)

type Middleware func(usersvc.Service) usersvc.Service

func subj(ctx context.Context) uuid.UUID {
	return ctx.Value(introspector.SubjectContextKey).(uuid.UUID)
}