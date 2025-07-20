package domain

type User struct {
	Id       int64
	WxOpenId string // 来区分不同的用户
	Username string
	Amount   int64
}
