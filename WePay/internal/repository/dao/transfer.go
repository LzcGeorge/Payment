package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type TransferDao interface {
	CreateTransferRequestRecord(ctx context.Context, req *TransferRequestRecord) error
	UpdateTransferRequestStatus(ctx context.Context, outbillno string, status string) error
	GetTransferStatus(ctx context.Context, outbillno string) (string, error)
	GetTransferRecord(ctx context.Context, outbillno string) (TransferRequestRecord, error)
}

type TransferRequestRecord struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	OutBillNo string
	Openid    string
	MchId     string
	Amount    int64
	Remark    string
	SceneId   string
	Status    string
	Ctime     time.Time
	Utime     time.Time
}

type GormTransferDao struct {
	db *gorm.DB
}

func NewTransferDao(db *gorm.DB) TransferDao {
	return &GormTransferDao{db: db}
}

func (d *GormTransferDao) CreateTransferRequestRecord(ctx context.Context, req *TransferRequestRecord) error {
	return d.db.Create(req).Error
}

// UpdateTransferRequestStatus 修改 Status
func (d *GormTransferDao) UpdateTransferRequestStatus(ctx context.Context, outbillno string, status string) error {
	return d.db.Model(&TransferRequestRecord{}).Where("out_bill_no = ?", outbillno).Updates(
		map[string]interface{}{
			"status": status,
			"utime":  time.Now(),
		},
	).Error
}

func (d *GormTransferDao) GetTransferStatus(ctx context.Context, outbillno string) (string, error) {
	var status string
	err := d.db.Model(&TransferRequestRecord{}).Where("out_bill_no = ?", outbillno).Select("status").Scan(&status).Error
	return status, err
}

func (d *GormTransferDao) GetTransferRecord(ctx context.Context, outbillno string) (TransferRequestRecord, error) {
	var record TransferRequestRecord
	err := d.db.Model(&TransferRequestRecord{}).Where("out_bill_no = ?", outbillno).First(&record).Error
	return record, err
}
