package service

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

func (s *inmemService) CreateUser(c context.Context, user User) (error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, u := range s.m {
		if u.Email == user.Email {
			return ErrUserExists
		}
	}
	s.m[user.ID] = &user
	return nil
}

func (s *inmemService) UpdateUser(c context.Context, user User) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.m[user.ID] = &user
	return nil
}

func (s *inmemService) Authenticate(email string, password string) (User, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	email = strings.ToLower(email)
	for id, u := range s.m {
		if u.Email == email {
			err := bcrypt.CompareHashAndPassword(s.p[id], []byte(password))
			if err == nil {
				return *s.m[id], nil
			}
			return User{}, ErrWrongPassword
		}
	}
	return User{}, ErrWrongEmail
}

func (s *inmemService) SetPassword(c context.Context, userId uuid.UUID, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.p[userId] = hashed
	return nil
}
