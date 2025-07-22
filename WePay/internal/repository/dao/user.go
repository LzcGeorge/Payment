package dao

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	WxOpenId string `gorm:"uniqueIndex;type:varchar(128)"`
	Username string
	Balance  int64
}

type UserDao interface {
	GetAmount(ctx context.Context, openid string) (int64, error)
	UpsertBalance(ctx context.Context, openid string, amount int64) error
}

type GormUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GormUserDao{db: db}
}

func (d *GormUserDao) GetAmount(ctx context.Context, openid string) (int64, error) {
	var user User
	err := d.db.Where("wx_open_id = ?", openid).First(&user).Error
	return user.Balance, err
}

func (d *GormUserDao) UpsertBalance(ctx context.Context, openid string, amount int64) error {
	user := User{
		WxOpenId: openid,
		Username: openid,
		Balance:  amount,
	}
	return d.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "wx_open_id"}}, // 依据 openid 冲突
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": gorm.Expr("balance + ?", amount), // 累加
			}),
		}).
		Create(&user).Error
}
