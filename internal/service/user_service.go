package service

import (
	"context"
	"fmt"
	"sync"

	"SIP/internal/model"
	"SIP/internal/repository"
	"SIP/internal/util"
)

type UserService struct {
	userRepo repository.UserRepository
	idGen    *util.IDGenerator
}

var (
	userServiceOnce sync.Once
	userServiceInst *UserService
)

func NewUserService(userRepo repository.UserRepository, idGen *util.IDGenerator) *UserService {
	userServiceOnce.Do(func() {
		userServiceInst = &UserService{userRepo: userRepo, idGen: idGen}
	})
	return userServiceInst
}

func (s *UserService) CreateUser(ctx context.Context, userID, name string) (model.User, error) {
	if name == "" {
		return model.User{}, fmt.Errorf("%w: name is required", util.ErrValidation)
	}
	if userID == "" {
		userID = s.idGen.Next("user")
	}
	user := model.User{UserID: userID, Name: name}
	return s.userRepo.Create(ctx, user)
}

func (s *UserService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.userRepo.List(ctx)
}
