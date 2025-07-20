package service

type TransferToUserResponse struct {
	OutBillNo      *string             `json:"out_bill_no,omitempty"`
	TransferBillNo *string             `json:"transfer_bill_no,omitempty"`
	CreateTime     *string             `json:"create_time,omitempty"`
	State          *TransferBillStatus `json:"state,omitempty"`
	PackageInfo    *string             `json:"package_info,omitempty"`
}

type TransferBillStatus string

func (e TransferBillStatus) Ptr() *TransferBillStatus {
	return &e
}

const (
	TRANSFERBILLSTATUS_ACCEPTED          TransferBillStatus = "ACCEPTED"
	TRANSFERBILLSTATUS_PROCESSING        TransferBillStatus = "PROCESSING"
	TRANSFERBILLSTATUS_WAIT_USER_CONFIRM TransferBillStatus = "WAIT_USER_CONFIRM"
	TRANSFERBILLSTATUS_TRANSFERING       TransferBillStatus = "TRANSFERING"
	TRANSFERBILLSTATUS_SUCCESS           TransferBillStatus = "SUCCESS"
	TRANSFERBILLSTATUS_FAIL              TransferBillStatus = "FAIL"
	TRANSFERBILLSTATUS_CANCELING         TransferBillStatus = "CANCELING"
	TRANSFERBILLSTATUS_CANCELLED         TransferBillStatus = "CANCELLED"
)

type TransferSceneReportInfo struct {
	InfoType    *string `json:"info_type,omitempty"`
	InfoContent *string `json:"info_content,omitempty"`
}

type TransferToUserRequest struct {
	Appid                    *string                   `json:"appid,omitempty"`
	OutBillNo                *string                   `json:"out_bill_no,omitempty"`
	TransferSceneId          *string                   `json:"transfer_scene_id,omitempty"`
	Openid                   *string                   `json:"openid,omitempty"`
	MchId                    *string                   `json:"mch_id,omitempty"`
	UserName                 *string                   `json:"user_name,omitempty"`
	TransferAmount           *int64                    `json:"transfer_amount,omitempty"`
	TransferRemark           *string                   `json:"transfer_remark,omitempty"`
	NotifyUrl                *string                   `json:"notify_url,omitempty"`
	UserRecvPerception       *string                   `json:"user_recv_perception,omitempty"`
	TransferSceneReportInfos []TransferSceneReportInfo `json:"transfer_scene_report_infos,omitempty"`
}
