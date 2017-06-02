package classsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/google/uuid"
	"github.com/studiously/classsvc/models"
)

func InstrumentingMiddleware(
	requestCount metrics.Counter,
	requestLatency metrics.Histogram,
) Middleware {
	return func(next Service) Service {
		return instrumentingMiddleware{requestCount, requestLatency, next}
	}
}

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

func (im instrumentingMiddleware) ListClasses(ctx context.Context) (classes []uuid.UUID, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "ListClasses", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.ListClasses(ctx)
}

func (im instrumentingMiddleware) GetClass(ctx context.Context, classID uuid.UUID) (class *models.Class, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetClass", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.GetClass(ctx, classID)
}

func (im instrumentingMiddleware) CreateClass(ctx context.Context, name string) (classID *uuid.UUID, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "CreateClass", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.CreateClass(ctx, name)
}

func (im instrumentingMiddleware) UpdateClass(ctx context.Context, classID uuid.UUID, name *string, currentUnit *uuid.UUID) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "UpdateClass", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.UpdateClass(ctx, classID, name, currentUnit)
}

func (im instrumentingMiddleware) DeleteClass(ctx context.Context, classID uuid.UUID) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "DeleteClass", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.DeleteClass(ctx, classID)
}

func (im instrumentingMiddleware) JoinClass(ctx context.Context, classID uuid.UUID) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "JoinClass", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.JoinClass(ctx, classID)
}

func (im instrumentingMiddleware) LeaveClass(ctx context.Context, userID *uuid.UUID, classID uuid.UUID) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "LeaveClass", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.LeaveClass(ctx, userID, classID)
}

func (im instrumentingMiddleware) SetRole(ctx context.Context, classID uuid.UUID, userID uuid.UUID, role models.UserRole) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "SetRole", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.SetRole(ctx, classID, userID, role)
}

func (im instrumentingMiddleware) ListMembers(ctx context.Context, classID uuid.UUID) (members []*models.Member, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "ListMembers", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.ListMembers(ctx, classID)
}

func (im instrumentingMiddleware) GetMember(ctx context.Context, classID, userID uuid.UUID) (member *models.Member, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetMember", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return im.next.GetMember(ctx, classID, userID)
}
