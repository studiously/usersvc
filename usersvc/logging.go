package usersvc

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/ory/hydra/oauth2"
	"github.com/studiously/introspector"
	"github.com/studiously/usersvc/models"
)

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (lm loggingMiddleware) GetProfile(ctx context.Context, userID uuid.UUID) (name string, err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "GetProfile",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"target", userID.String(),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.GetProfile(ctx, userID)
}

func (lm loggingMiddleware) GetUserInfo(ctx context.Context) (user *models.User, err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "GetUserInfo",
			"user", subj(ctx).String(),
			"client", cli(ctx),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.GetUserInfo(ctx)
}

func (lm loggingMiddleware) CreateUser(name, email, password string) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "CreateUser",
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.CreateUser(name, email, password)
}

func (lm loggingMiddleware) SetName(ctx context.Context, name string) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "SetName",
			"user", subj(ctx),
			"client", cli(ctx),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.SetName(ctx, name)
}

func (lm loggingMiddleware) SetEmail(ctx context.Context, email string) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "SetEmail",
			"user", subj(ctx),
			"client", cli(ctx),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.SetEmail(ctx, email)
}

func (lm loggingMiddleware) SetPassword(ctx context.Context, password string) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "SetPassword",
			"user", subj(ctx),
			"client", cli(ctx),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.SetPassword(ctx, password)
}

func (lm loggingMiddleware) Authenticate(email string, password string) (user uuid.UUID, err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "Authenticate",
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.Authenticate(email, password)
}

func (lm loggingMiddleware) DeleteUser(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"action", "DeleteUser",
			"user", subj(ctx),
			"duration", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return lm.next.DeleteUser(ctx)
}

func cli(ctx context.Context) string {
	return ctx.Value(introspector.OAuth2IntrospectionContextKey).(oauth2.Introspection).ClientID
}
