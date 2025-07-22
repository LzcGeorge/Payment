package domain

type TransferRecord struct {
	ID          int64
	OutBillNo   string // 转账单号
	Openid      string // 转账用户ID
	MchId       string // 商户ID
	Amount      int64  // 转账金额
	Remark      string // 转账备注
	SceneId     string // 转账场景ID
	Status      string // 转账状态
	PackageInfo string // notify 的时候用
}

const (
	TransferStatusAccepted        = "ACCEPTED"
	TransferStatusProcessing      = "PROCESSING"
	TransferStatusWaitUserConfirm = "WAIT_USER_CONFIRM"
	TransferStatusTransfering     = "TRANSFERING"
	TransferStatusSuccess         = "SUCCESS"
	TransferStatusFail            = "FAIL"
	TransferStatusCanceling       = "CANCELING"
	TransferStatusCancelled       = "CANCELLED"
)
