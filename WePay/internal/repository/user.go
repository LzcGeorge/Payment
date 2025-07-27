package repository

import (
	"context"
	"wepay/internal/domain"
	"wepay/internal/repository/dao"
)

type UserRepository interface {
	GetAmount(ctx context.Context, openid string) (int64, error)
	UpdateBalance(ctx context.Context, openid string, amount int64) error
	GetUser(ctx context.Context, openid string) (*domain.User, error)
	InsertUser(ctx context.Context, openid string) error
}

type userRepository struct {
	dao dao.UserDao
}

func NewUserRepository(dao dao.UserDao) UserRepository {
	return &userRepository{dao: dao}
}

func (r *userRepository) GetAmount(ctx context.Context, openid string) (int64, error) {
	return r.dao.GetAmount(ctx, openid)
}

func (r *userRepository) UpdateBalance(ctx context.Context, openid string, amount int64) error {
	return r.dao.UpsertBalance(ctx, openid, amount)
}

func (r *userRepository) GetUser(ctx context.Context, openid string) (*domain.User, error) {
	user, err := r.dao.GetUser(ctx, openid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	return &domain.User{
		WxOpenId: user.WxOpenId,
		Amount:   user.Balance,
	}, nil
}

func (r *userRepository) InsertUser(ctx context.Context, openid string) error {
	return r.dao.InsertUser(ctx, openid)
}
