package usersvc

import (
	"context"
	"database/sql"
	"github.com/Studiously/usersvc/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func NewPersistentService(db *sql.DB) Service {
	return &persistentService{db}
}

type persistentService struct {
	*sql.DB
}

func (s *persistentService) GetUser(c context.Context, id uuid.UUID) (User, error) {
	u, err := models.UserByID(s.DB, id)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:   u.ID,
		Name: u.Name,
	}, nil
}

func (s *persistentService) GetProfile(c context.Context, id uuid.UUID) (Profile, error) {
	u, err := models.UserByID(s.DB, id)
	if err != nil {
		return Profile{}, err
	}
	return Profile{
		Name: u.Name,
	}, nil
}

func (s *persistentService) CreateUser(c context.Context, user User) (err error) {
	if u, _ := models.UserByEmail(s.DB, user.Email); u != nil && u.Exists() {
		return ErrUserExists
	}
	u := &models.User{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Active: user.Active,
	}
	err = u.Insert(s.DB)
	return
}

func (s *persistentService) UpdateUser(c context.Context, user User) error {
	u := &models.User{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Active: user.Active,
	}
	err := u.Upsert(s.DB)
	return err
}

func (s *persistentService) SetPassword(c context.Context, userId uuid.UUID, password string) error {
	li, err := models.LocalIdentityByUserID(s.DB, userId)
	if err != nil {
		li = &models.LocalIdentity{
			UserID: userId,
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ErrHashFailed
	}
	li.Password = string(hashed)
	err = li.Upsert(s.DB)
	return err
}

func (s *persistentService) Authenticate(c context.Context, email string, password string) (user User, err error) {
	u, err := models.UserByEmail(s.DB, email)
	if err != nil {
		return
	}
	if !u.Exists() {
		return User{}, ErrWrongEmail
	}
	li, err := models.LocalIdentityByUserID(s.DB, user.ID)
	if err != nil {
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(li.Password), []byte(password))
	if err != nil {
		return User{}, ErrWrongPassword
	}
	return User{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Active: u.Active,
	}, nil
}
