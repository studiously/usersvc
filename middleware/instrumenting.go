package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/google/uuid"
	"github.com/studiously/usersvc/models"
	"github.com/studiously/usersvc/usersvc"
)

func Instrumenting(
	requestCount metrics.Counter,
	requestLatency metrics.Histogram,
) Middleware {
	return func(next usersvc.Service) usersvc.Service {
		return instrumentingMiddleware{requestCount, requestLatency, next}
	}
}

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           usersvc.Service
}

func (im instrumentingMiddleware) GetProfile(ctx context.Context, userID uuid.UUID) (name string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetProfile", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.GetProfile(ctx, userID)
}

func (im instrumentingMiddleware) GetUserInfo(ctx context.Context) (user *models.User, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetUserInfo", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.GetUserInfo(ctx)
}

func (im instrumentingMiddleware) CreateUser(name, email, password string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "CreateUser", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.CreateUser(name, email, password)
}

func (im instrumentingMiddleware) SetName(ctx context.Context, name string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "SetName", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.SetName(ctx, name)
}

func (im instrumentingMiddleware) SetEmail(ctx context.Context, email string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "SetEmail", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.SetEmail(ctx, email)
}

func (im instrumentingMiddleware) SetPassword(ctx context.Context, password string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "SetPassword", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.SetPassword(ctx, password)
}

func (im instrumentingMiddleware) Authenticate(email string, password string) (userID uuid.UUID, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Authenticate", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.Authenticate(email, password)
}

func (im instrumentingMiddleware) DeleteUser(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "DeleteUser", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.DeleteUser(ctx)
}

func (im instrumentingMiddleware) ResetPassword(ctx context.Context, email string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "ResetPassword", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.ResetPassword(ctx, email)
}
