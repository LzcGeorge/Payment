package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type TransferRequest struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	OutBillNo string
	Openid    string
	Amount    int64
	Remark    string
	SceneId   string
	Status    int // 0: INIT, 1: SUCCESS, 2: FAIL
	Ctime     time.Time
	Utime     time.Time
}

type TransferDao struct {
	db *gorm.DB
}

func NewTransferDao(db *gorm.DB) *TransferDao {
	return &TransferDao{db: db}
}

func (d *TransferDao) CreateTransferRequest(ctx context.Context, req *TransferRequest) error {
	return d.db.Create(req).Error
}

// UpdateTransferRequestStatus 修改 Status
func (d *TransferDao) UpdateTransferRequestStatus(ctx context.Context, outbillno string, status int) error {
	return d.db.Model(&TransferRequest{}).Where("out_bill_no = ?", outbillno).Updates(
		map[string]interface{}{
			"status": status,
			"utime":  time.Now(),
		},
	).Error
}
