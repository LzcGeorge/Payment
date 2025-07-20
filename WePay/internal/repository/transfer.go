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

func (r *TransferRepository) CreateTransferRequest(ctx context.Context, req *domain.TransferRecord) error {
	return r.dao.CreateTransferRequestRecord(ctx, &dao.TransferRequestRecord{
		OutBillNo: req.OutBillNo,
		Openid:    req.Openid,
		MchId:     req.MchId,
		Amount:    req.Amount,
		Remark:    req.Remark,
		SceneId:   req.SceneId,
		Status:    req.Status,
		Ctime:     time.Now(),
		Utime:     time.Now(),
	})
}

func (r *TransferRepository) UpdateTransferRequestStatus(ctx context.Context, outbillno, state string) error {
	return r.dao.UpdateTransferRequestStatus(ctx, outbillno, state)
}

func (r *TransferRepository) GetTransferStatus(ctx context.Context, outbillno string) (string, error) {
	return r.dao.GetTransferStatus(ctx, outbillno)
}

func (r *TransferRepository) GetTransferRecord(ctx context.Context, outbillno string) (domain.TransferRecord, error) {
	record, err := r.dao.GetTransferRecord(ctx, outbillno)
	if err != nil {
		return domain.TransferRecord{}, err
	}
	return domain.TransferRecord{
		OutBillNo: record.OutBillNo,
		Openid:    record.Openid,
		Amount:    record.Amount,
		MchId:     record.MchId,
		Remark:    record.Remark,
		SceneId:   record.SceneId,
		Status:    record.Status,
	}, nil
}
