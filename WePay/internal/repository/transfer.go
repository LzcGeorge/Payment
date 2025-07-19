package repository

import (
	"context"
	"time"
	"wepay/internal/domain"
	"wepay/internal/repository/dao"
)

type TransferRepository struct {
	dao *dao.TransferDao
}

func NewTransferRepository(dao *dao.TransferDao) *TransferRepository {
	return &TransferRepository{dao: dao}
}

func (r *TransferRepository) CreateTransferRequest(ctx context.Context, req *domain.TransferRequest) error {
	return r.dao.CreateTransferRequest(ctx, &dao.TransferRequest{
		OutBillNo: req.OutBillNo,
		Openid:    req.Openid,
		Amount:    req.Amount,
		Remark:    req.Remark,
		SceneId:   req.SceneId,
		Status:    req.Status,
		Ctime:     time.Now(),
		Utime:     time.Now(),
	})
}
