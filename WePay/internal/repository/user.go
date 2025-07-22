package repository

import (
	"context"
	"wepay/internal/repository/dao"
)

type UserRepository interface {
	GetAmount(ctx context.Context, openid string) (int64, error)
	UpdateBalance(ctx context.Context, openid string, amount int64) error
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
