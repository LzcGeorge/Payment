package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

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

type TransferDao struct {
	db *gorm.DB
}

func NewTransferDao(db *gorm.DB) *TransferDao {
	return &TransferDao{db: db}
}

func (d *TransferDao) CreateTransferRequestRecord(ctx context.Context, req *TransferRequestRecord) error {
	return d.db.Create(req).Error
}

// UpdateTransferRequestStatus 修改 Status
func (d *TransferDao) UpdateTransferRequestStatus(ctx context.Context, outbillno string, status string) error {
	return d.db.Model(&TransferRequestRecord{}).Where("out_bill_no = ?", outbillno).Updates(
		map[string]interface{}{
			"status": status,
			"utime":  time.Now(),
		},
	).Error
}

func (d *TransferDao) GetTransferStatus(ctx context.Context, outbillno string) (string, error) {
	var status string
	err := d.db.Model(&TransferRequestRecord{}).Where("out_bill_no = ?", outbillno).Select("status").Scan(&status).Error
	return status, err
}

func (d *TransferDao) GetTransferRecord(ctx context.Context, outbillno string) (TransferRequestRecord, error) {
	var record TransferRequestRecord
	err := d.db.Model(&TransferRequestRecord{}).Where("out_bill_no = ?", outbillno).First(&record).Error
	return record, err
}
