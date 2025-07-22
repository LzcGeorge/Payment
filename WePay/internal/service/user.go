package service

import (
	"context"
	"wepay/internal/repository"
)

type UserService interface {
	GetAmount(ctx context.Context, openid string) (int64, error)
	UpdateBalance(ctx context.Context, openid string, amount int64) error
}

type userService struct {
	repo repository.UserRepository
	// NewUserService creates and returns a new UserService instance initialized with the provided UserRepository.
	// The UserService depends on the UserRepository for data access operations.
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetAmount(ctx context.Context, openid string) (int64, error) {
	return s.repo.GetAmount(ctx, openid)
}

func (s *userService) UpdateBalance(ctx context.Context, openid string, amount int64) error {
	return s.repo.UpdateBalance(ctx, openid, amount)
}
