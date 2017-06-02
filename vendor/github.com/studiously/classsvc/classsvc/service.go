package classsvc

import (
	"context"

	"github.com/google/uuid"
	"github.com/studiously/classsvc/models"
)

type Middleware func(Service) Service

type Service interface {
	// ListClasses gets all classes the current user is enrolled in.
	ListClasses(ctx context.Context) ([]uuid.UUID, error)
	// GetClass gets details for a specific class.
	GetClass(ctx context.Context, classID uuid.UUID) (*models.Class, error)
	// CreateClass creates a class and enrolls the current user in it as an administrator.
	CreateClass(ctx context.Context, name string) (*uuid.UUID, error)
	// UpdateClass updates a class.
	UpdateClass(ctx context.Context, classID uuid.UUID, name *string, currentUnit *uuid.UUID) error
	// DeleteClass deactivates a class.
	DeleteClass(ctx context.Context, classID uuid.UUID) error
	// JoinClass enrolls the current user in a class.
	JoinClass(ctx context.Context, classID uuid.UUID) (error)
	// LeaveClass causes a user to be un-enrolled from a class.
	// If user is not `uuid.Nil`, then LeaveClass removes the other user, requiring the current user to have elevated permissions.
	LeaveClass(ctx context.Context, userID *uuid.UUID, classID uuid.UUID) error
	// SetRole sets the role of a user in a class.
	// The current user must have a higher role than the target user.
	SetRole(ctx context.Context, classID uuid.UUID, userID uuid.UUID, role models.UserRole) error
	// ListMembers lists all members of a class and their role.
	ListMembers(ctx context.Context, classID uuid.UUID) ([]*models.Member, error)
	////GetMember gets a member of a class.
	GetMember(ctx context.Context, classID, userID uuid.UUID) (member *models.Member, err error)
	//GetRole(ctx context.Context, userID, classID uuid.UUID) (*models.UserRole, error)
	//IsOwner(ctx context.Context, userID, classID uuid.UUID) (bool, error)
}
