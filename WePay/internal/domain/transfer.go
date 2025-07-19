package domain

// domain/transfer_request.go
type TransferRequest struct {
	ID        int64
	OutBillNo string
	Openid    string
	Amount    int64
	Remark    string
	SceneId   string
	Status    int // 0: INIT, 1: SUCCESS, 2: FAIL
}

const (
	StatusInit = iota
	StatusSuccess
	StatusFail
)
