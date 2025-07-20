package repository

import (
	"context"
	"wepay/internal/repository/dao"
)

type UserRepository struct {
	dao *dao.UserDao
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{dao: dao}
}

func (r *UserRepository) GetAmount(ctx context.Context, openid string) (int64, error) {
	return r.dao.GetAmount(ctx, openid)
}

func (r *UserRepository) UpdateBalance(ctx context.Context, openid string, amount int64) error {
	return r.dao.UpsertBalance(ctx, openid, amount)
}
