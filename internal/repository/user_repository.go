package repository

import (
	"context"
	"fmt"
	"sync"

	"SIP/internal/model"
	"SIP/internal/util"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	GetByID(ctx context.Context, userID string) (model.User, error)
	List(ctx context.Context) ([]model.User, error)
}

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]model.User
}

var (
	userRepositoryOnce sync.Once
	userRepositoryInst *InMemoryUserRepository
)

func NewInMemoryUserRepository(seed []model.User) *InMemoryUserRepository {
	userRepositoryOnce.Do(func() {
		m := make(map[string]model.User, len(seed))
		for _, u := range seed {
			m[u.UserID] = u
		}
		userRepositoryInst = &InMemoryUserRepository{users: m}
	})
	return userRepositoryInst
}

func (r *InMemoryUserRepository) Create(_ context.Context, user model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.UserID]; ok {
		return model.User{}, fmt.Errorf("%w: user %s already exists", util.ErrConflict, user.UserID)
	}
	r.users[user.UserID] = user
	return user, nil
}

func (r *InMemoryUserRepository) GetByID(_ context.Context, userID string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[userID]
	if !ok {
		return model.User{}, fmt.Errorf("%w: user %s", util.ErrNotFound, userID)
	}
	return u, nil
}

func (r *InMemoryUserRepository) List(_ context.Context) ([]model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	return out, nil
}
