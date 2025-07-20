package service

import (
	"context"
	"wepay/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
	// NewUserService creates and returns a new UserService instance initialized with the provided UserRepository.
	// The UserService depends on the UserRepository for data access operations.
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAmount(ctx context.Context, openid string) (int64, error) {
	return s.repo.GetAmount(ctx, openid)
}

func (s *UserService) UpdateBalance(ctx context.Context, openid string, amount int64) error {
	return s.repo.UpdateBalance(ctx, openid, amount)
}
