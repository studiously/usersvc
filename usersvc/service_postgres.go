package usersvc

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/studiously/classsvc/classsvc"
	"github.com/studiously/introspector"
	"github.com/studiously/usersvc/models"
	"golang.org/x/crypto/bcrypt"
)

func New(db *sql.DB, cs classsvc.Service) Service {
	return &postgresService{
		db,
		cs,
	}
}

type postgresService struct {
	*sql.DB
	cs classsvc.Service
}

func (s *postgresService) GetProfile(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := models.UserByID(s.DB, userID)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (s *postgresService) GetUserInfo(ctx context.Context) (*models.User, error) {
	return models.UserByID(s, subj(ctx))
}

func (s *postgresService) CreateUser(name, email, password string) (err error) {
	if u, _ := models.UserByEmail(s, email); u != nil && u.Exists() {
		return ErrUserExists
	}
	tx, err := s.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
	u := &models.User{
		ID:     uuid.New(),
		Name:   name,
		Email:  email,
		Active: true,
	}
	err = u.Insert(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	li, err := models.LocalIdentityByUserID(s.DB, u.ID)
	if err != nil {
		li = &models.LocalIdentity{
			UserID: u.ID,
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return ErrHashFailed
	}
	li.Password = string(hashed)
	err = li.Upsert(s)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *postgresService) SetName(ctx context.Context, name string) error {
	user, err := models.UserByID(s, subj(ctx))
	if err != nil {
		return err
	}
	user.Name = name
	return user.Update(s)
}

func (s *postgresService) SetEmail(ctx context.Context, email string) error {
	user, err := models.UserByID(s, subj(ctx))
	if err != nil {
		return err
	}
	user.Email = email
	return user.Update(s)
}

func (s *postgresService) SetPassword(ctx context.Context, password string) error {
	li, err := models.LocalIdentityByUserID(s.DB, subj(ctx))
	if err != nil {
		li = &models.LocalIdentity{
			UserID: subj(ctx),
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ErrHashFailed
	}
	li.Password = string(hashed)
	err = li.Upsert(s)
	return err
}

func (s *postgresService) Authenticate(email string, password string) (userID uuid.UUID, err error) {
	u, err := models.UserByEmail(s.DB, email)
	if err != nil {
		return
	}
	if !u.Exists() {
		return uuid.Nil, ErrWrongEmail
	}
	li, err := models.LocalIdentityByUserID(s.DB, u.ID)
	if err != nil {
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(li.Password), []byte(password))
	if err != nil {
		return uuid.Nil, ErrWrongPassword
	}
	return u.ID, nil
}

func (s *postgresService) DeleteUser(ctx context.Context) error {
	u, err := models.UserByID(s, subj(ctx))
	if err != nil {
		return err
	}
	if !u.Exists() {
		return ErrNotFound
	}

	classes, err := s.cs.ListClasses(ctx)
	if err != nil {
		return err
	}

	for _, class := range classes {
		member, err := s.cs.GetMember(ctx, class, subj(ctx))
		if err != nil {
			return err
		}
		if member.Owner {
			return ErrDeleteOwner
		}
	}

	u.Active = false
	return u.Save(s)
}

func (s *postgresService) ResetPassword(ctx context.Context, email string) error {
	//u, err := models.UserByEmail(s, email)
	//switch err {
	//case nil:
	//	break
	//case sql.ErrNoRows:
	//	return nil
	//default:
	//	return err
	//}
	panic("unimplemented")
	return nil
}

func subj(ctx context.Context) uuid.UUID {
	return ctx.Value(introspector.SubjectContextKey).(uuid.UUID)
}
