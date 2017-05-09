package usersvc

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrUserExists      = errors.New("user exists")
	ErrWrongEmail      = errors.New("wrong email")
	ErrWrongPassword   = errors.New("wrong password")
	ErrUnauthenticated = errors.New("must be authenticated")
)

type Service interface {
	GetUser(c context.Context, id uuid.UUID) (User, error)
	GetProfile(c context.Context, id uuid.UUID) (Profile, error)
	CreateUser(c context.Context, user User) (id uuid.UUID, err error)
	UpdateUser(c context.Context, user User) error
	SetPassword(c context.Context, userId uuid.UUID, password []byte) error
	Authenticate(c context.Context, email string, password []byte) (User, error)
}

type User struct {
	Id     uuid.UUID `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool `json:"active"`
}

type Profile struct {
	Name string `json:"name"`
}

type inmemService struct {
	m   map[uuid.UUID]*User
	p   map[uuid.UUID][]byte
	mtx sync.RWMutex
}

func NewInmemService() Service {
	//wid, err := strconv.Atoi(os.Getenv("WORKER_ID"))
	//if err != nil || wid == 0 {
	//	// WID isn't necessary for in-memory, use random
	//	wid = rand.Int() % 1024
	//}
	//sf, err := snowflake.New(uint32(wid))
	//if err != nil {
	//	logrus.WithError(err).Fatal("failed to initialize snowflake worker")
	//}
	return &inmemService{
		m:   make(map[uuid.UUID]*User),
		p:   make(map[uuid.UUID][]byte),
		mtx: sync.RWMutex{},
	}
}

func (s *inmemService) GetUser(c context.Context, id uuid.UUID) (User, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if u, ok := s.m[id]; ok {
		return *u, nil
	}
	return User{}, ErrNotFound
}

func (s *inmemService) GetProfile(c context.Context, id uuid.UUID) (Profile, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if u, ok := s.m[id]; ok {
		return Profile{
			Name: u.Name,
		}, nil
	}
	return Profile{}, ErrNotFound
}

func (s *inmemService) CreateUser(c context.Context, user User) (uuid.UUID, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	user.Email = strings.ToLower(user.Email)
	for _, u := range s.m {
		if u.Email == user.Email {
			return uuid.Nil, ErrUserExists
		}
	}
	id := uuid.NewV4()
	s.m[id] = &User{
		Id:     id,
		Name:   user.Name,
		Email:  user.Email,
		Active: user.Active,
	}
	return id, nil
}

func (s *inmemService) UpdateUser(c context.Context, user User) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.m[user.Id] = &user
	return nil
}

func (s *inmemService) Authenticate(c context.Context, email string, password []byte) (User, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	email = strings.ToLower(email)
	for id, u := range s.m {
		if u.Email == email {
			err := bcrypt.CompareHashAndPassword(s.p[id], password)
			if err == nil {
				return *s.m[id], nil
			}
			return User{}, ErrWrongPassword
		}
	}
	return User{}, ErrWrongEmail
}

func (s *inmemService) SetPassword(c context.Context, userId uuid.UUID, password []byte) error {
	hashed, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.p[userId] = hashed
	return nil
}
