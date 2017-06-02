package classsvc

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/ory/hydra/oauth2"
	"github.com/studiously/classsvc/models"
	"github.com/studiously/introspector"
)

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (lm loggingMiddleware) ListClasses(ctx context.Context) (classes []uuid.UUID, err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "ListClasses",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.ListClasses(ctx)
}

func (lm loggingMiddleware) GetClass(ctx context.Context, classID uuid.UUID) (classes *models.Class, err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "GetClass",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.GetClass(ctx, classID)
}

func (lm loggingMiddleware) CreateClass(ctx context.Context, name string) (classID *uuid.UUID, err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "CreateClass",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.CreateClass(ctx, name)
}

func (lm loggingMiddleware) UpdateClass(ctx context.Context, classID uuid.UUID, name *string, currentUnit *uuid.UUID) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "UpdateClass",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.UpdateClass(ctx, classID, name, currentUnit)
}

func (lm loggingMiddleware) DeleteClass(ctx context.Context, classID uuid.UUID) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "DeleteClass",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.DeleteClass(ctx, classID)
}

func (lm loggingMiddleware) JoinClass(ctx context.Context, classID uuid.UUID) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "JoinClass",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.JoinClass(ctx, classID)
}

func (lm loggingMiddleware) LeaveClass(ctx context.Context, userID *uuid.UUID, classID uuid.UUID) (err error) {
	defer func(begin time.Time) {
		target := userID
		if target == nil {
			subj := subj(ctx)
			target = &subj
		}
		lm.logger.Log(
			"action", "LeaveClass",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"target", target.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.LeaveClass(ctx, userID, classID)
}

func (lm loggingMiddleware) SetRole(ctx context.Context, classID uuid.UUID, userID uuid.UUID, role models.UserRole) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "SetRole",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"target", userID.String(),
			"role", role.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.SetRole(ctx, classID, userID, role)
}

func (lm loggingMiddleware) ListMembers(ctx context.Context, classID uuid.UUID) (members []*models.Member, err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "ListMembers",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.ListMembers(ctx, classID)
}

func (lm loggingMiddleware) GetMember(ctx context.Context, classID, userID uuid.UUID) (member *models.Member, err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "GetMember",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"class", classID.String(),
			"target", userID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.GetMember(ctx, classID, userID)
}

func cli(ctx context.Context) string {
	return ctx.Value(introspector.OAuth2IntrospectionContextKey).(oauth2.Introspection).ClientID
}
