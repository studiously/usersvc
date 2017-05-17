package service

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestInmemService_CreateUser(t *testing.T) {
	service := newInmem()
	user := newUser()
	service.CreateUser(user)
	if _, ok := service.m[user.ID.String()]; !ok {
		t.Error("User was not saved.")
	}
}

func TestInmemService_GetUser(t *testing.T) {
	service := newInmem()
	user := newUser()
	service.m[user.ID.String()] = &user

	ru, err := service.GetUser(user.ID)
	if err != nil {
		t.Errorf("Failed to retrieve user: %v", err)
		return
	}
	if ru.ID.String() != user.ID.String() {
		t.Error("Returned user does not match original.")
	}
}

func TestInmemService_GetProfile(t *testing.T) {
	service := newInmem()
	user := newUser()
	service.m[user.ID.String()] = &user

	if profile, err := service.GetProfile(user.ID); err != nil || profile.Name != user.Name {
		t.Error("Profile is not correct or could not be retrieved.")
	}
}

func TestInmemService_UpdateUser(t *testing.T) {
	service := newInmem()
	user := newUser()
	service.CreateUser(user)

	user.Name = "Bonnie Flint"
	err := service.UpdateUser(user)
	if err != nil {
		t.Errorf("Failed to update user: %v", err)
		return
	}
	if service.m[user.ID.String()].Name != user.Name {
		t.Error("Failed to update user.")
	}
}

func TestInmemService_SetPassword(t *testing.T) {
	service := newInmem()
	id := uuid.New()

	service.SetPassword(id, "supersecret")

	if err := bcrypt.CompareHashAndPassword(service.p[id.String()], []byte("supersecret")); err != nil {
		t.Errorf("Failed to set password: %v", err)
	}
}

func TestInmemService_Authenticate(t *testing.T) {
	service := newInmem()
	user := newUser()

	err := service.CreateUser(user)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
		return
	}
	err = service.SetPassword(user.ID, "supersecret")
	if err != nil {
		t.Errorf("Failed to set password: %v", err)
		return
	}

	au, err := service.Authenticate("johnny.appleseed@example.com", "supersecret")
	if err != nil {
		t.Errorf("Failed to authenticate: %v", err)
		return
	}
	if au.ID.String() != user.ID.String() {
		t.Error("Authenticated user does not match original.")
	}
}

func newUser() User {
	return User{
		uuid.New(),
		"Johnny Appleseed",
		"johnny.appleseed@example.com",
		true,
	}
}

func newInmem() *inmemService {
	return &inmemService{
		m:   make(map[string]*User),
		p:   make(map[string][]byte),
		mtx: sync.RWMutex{},
	}
}
