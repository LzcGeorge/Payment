package dao

type User struct {
	Id       int64 `gorm:"primaryKey;autoIncrement"`
	WxOpenId string
	Username string
	Amount   int64
}
