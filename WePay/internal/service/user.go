package service

import (
	"context"
	"log"
	"wepay/internal/repository"
)

type UserService interface {
	GetAmount(ctx context.Context, openid string) (int64, error)
	UpdateBalance(ctx context.Context, openid string, amount int64) error
	UpsertUser(ctx context.Context, openid string) error
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

func (s *userService) UpsertUser(ctx context.Context, openid string) error {
	user, err := s.repo.GetUser(ctx, openid)
	log.Println("user", user)

	if user == nil {
		return s.repo.InsertUser(ctx, openid)
	}
	return err
}
