package store

import (
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

//User stores user's information
type User struct {
	UserName       string
	HashedPassword string
	Role           string
}

//UserStore stores the user dta
type UserStore interface {
	Save(user *User) error
	Find(userName string) (*User, error)
}

//InMemoryUserStore stores the user data
type InMemoryUserStore struct {
	mutex sync.RWMutex
	Users map[string]*User
}

//NewUser creates a new user on ths system
func NewUser(userName string, password string, role string) (*User, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	user := &User{
		HashedPassword: string(hashedPass),
		UserName:       userName,
		Role:           role,
	}

	return user, nil
}

//NewInMemoryUserStore is teh ctor
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		Users: make(map[string]*User),
	}
}

//Save saves th data t memory
func (store *InMemoryUserStore) Save(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.Users[user.UserName] != nil {
		return ErrAlreadyExists
	}

	store.Users[user.UserName] = user.Clone()
	return nil
}

//Find finds the user in teh map
func (store *InMemoryUserStore) Find(userName string) (*User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	user := store.Users[userName]
	if user == nil {
		return nil, nil
	}

	return user.Clone(), nil
}

//IsCorrectPassword verifies teh password
func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil
}

//Clone clones the data replica
func (user *User) Clone() *User {
	return &User{
		UserName:       user.UserName,
		HashedPassword: user.HashedPassword,
		Role:           user.Role,
	}
}
