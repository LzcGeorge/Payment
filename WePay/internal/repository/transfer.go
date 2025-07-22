package repository

import (
	"context"
	"time"
	"wepay/internal/domain"
	"wepay/internal/repository/dao"
)

type TransferRepository interface {
	CreateTransferRequest(ctx context.Context, req *domain.TransferRecord) error
	UpdateTransferRequestStatus(ctx context.Context, outbillno, state string) error
	GetTransferStatus(ctx context.Context, outbillno string) (string, error)
	GetTransferRecordByOutBillNo(ctx context.Context, outbillno string) (domain.TransferRecord, error)
	GetTransferRecordByPackageInfo(ctx context.Context, packageInfo string) (domain.TransferRecord, error)
}

type transferRepository struct {
	dao dao.TransferDao
}

func NewTransferRepository(dao dao.TransferDao) TransferRepository {
	return &transferRepository{dao: dao}
}

func (r *transferRepository) CreateTransferRequest(ctx context.Context, req *domain.TransferRecord) error {
	return r.dao.CreateTransferRequestRecord(ctx, &dao.TransferRequestRecord{
		OutBillNo:   req.OutBillNo,
		Openid:      req.Openid,
		MchId:       req.MchId,
		Amount:      req.Amount,
		Remark:      req.Remark,
		SceneId:     req.SceneId,
		Status:      req.Status,
		PackageInfo: req.PackageInfo,
		Ctime:       time.Now(),
		Utime:       time.Now(),
	})
}

func (r *transferRepository) UpdateTransferRequestStatus(ctx context.Context, outbillno, state string) error {
	return r.dao.UpdateTransferRequestStatus(ctx, outbillno, state)
}

func (r *transferRepository) GetTransferStatus(ctx context.Context, outbillno string) (string, error) {
	return r.dao.GetTransferStatus(ctx, outbillno)
}

func (r *transferRepository) GetTransferRecordByOutBillNo(ctx context.Context, outbillno string) (domain.TransferRecord, error) {
	record, err := r.dao.GetTransferRecordByOutBillNo(ctx, outbillno)
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

func (r *transferRepository) GetTransferRecordByPackageInfo(ctx context.Context, packageInfo string) (domain.TransferRecord, error) {
	record, err := r.dao.GetTransferRecordByPackageInfo(ctx, packageInfo)
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
