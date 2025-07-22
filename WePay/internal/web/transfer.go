package web

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"wepay/internal/domain"
	"wepay/internal/service"
	"wepay/internal/service/wxpay_utility"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
)

type TransferHandler struct {
	svc     service.TransferService
	userSvc service.UserService
	client  Client
}

func NewTransferHandler(svc service.TransferService, userSvc service.UserService, client Client) *TransferHandler {
	return &TransferHandler{
		svc:     svc,
		userSvc: userSvc,
		client:  client,
	}
}

// RegisterRoutes registers the HTTP routes for transfer operations on the provided router group.
// Currently, it registers a POST endpoint at "/to_user" that triggers the InitiateTransfer handler method.
func (t *TransferHandler) RegisterRoutes(ug *gin.RouterGroup) {
	ug.POST("/to_user", t.InitiateTransfer)
	ug.POST("/notify", t.TransferNotify)   // 微信支付的回调（手动模拟实现）
	ug.POST("/confirm", t.ConfirmTransfer) // 确认转账
	ug.GET("/amount", t.FetchAmount)       // 查询余额
}

func generatePackageInfo(openid string, timeStr string) string {
	return fmt.Sprintf("PK%s-%s", openid, timeStr)
}

func (t *TransferHandler) InitiateTransfer(ctx *gin.Context) {
	// 用户传来的参数
	var req struct {
		Openid string `form:"openid" json:"openid" binding:"required"`
		Amount int64  `form:"amount" json:"amount" binding:"required"`
		Remark string `json:"remark"`
		Time   string `json:"time" binding:"required"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法: " + err.Error()})
		return
	}
	openid := req.Openid // 用户openid
	amount := req.Amount // 转账金额，单位为 ”分“
	remark := req.Remark // 转账备注
	timeStr := req.Time  // 转账时间

	// 商户的配置 MchConfig, appid， transfer_scene_id（转账场景）, notify_url（通知URL）, user_recv_perception（用户收款码）
	transfer_scene_id := "1000" // 转账场景：现金营销

	notify_url := t.client.NotifyUrl
	user_recv_perception := "现金红包" // 用户收款时感知到的收款原因将根据转账场景自动展示默认内容。

	mchConfig, err := wxpay_utility.CreateMchConfig(
		t.client.Mchid,
		t.client.CertificateSerialNo,
		t.client.PrivateKeyPath,
		t.client.WechatPayPublicKeyId,
		t.client.WechatPayPublicKeyPath,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一outbillno并保存转账请求
	outbillno := t.svc.GenerateOutBillNo(openid, amount)
	packageInfo := generatePackageInfo(openid, timeStr)
	requestRecord := &domain.TransferRecord{
		OutBillNo:   outbillno,
		Openid:      openid,
		MchId:       t.client.Mchid,
		PackageInfo: packageInfo,
		Amount:      amount,
		Remark:      remark,
		Status:      domain.TransferStatusProcessing,
	}
	err = t.svc.AddTransferRequest(ctx, requestRecord)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构造 TransferToUserRequest
	request := &service.TransferToUserRequest{
		// 商家
		Appid:              core.String(t.client.Appid), // 小程序与商户关联的appid
		OutBillNo:          core.String(outbillno),
		TransferSceneId:    core.String(transfer_scene_id),
		Openid:             core.String(openid),
		MchId:              core.String(t.client.Mchid),
		UserName:           core.String(remark),
		TransferAmount:     core.Int64(amount),
		TransferRemark:     core.String(remark),
		NotifyUrl:          core.String(notify_url),
		UserRecvPerception: core.String(user_recv_perception),
	}

	// 发起转账
	_, err = t.svc.TransferToUser(mchConfig, request)
	if err != nil {
		log.Println("post to wx error:", err)
	}
	response := &service.TransferToUserResponse{
		OutBillNo:      core.String(outbillno),
		TransferBillNo: core.String("1330000071100999991182020050700019480001"),
		CreateTime:     core.String("2015-05-20T13:29:35.120+08:00"),
		State:          service.TRANSFERBILLSTATUS_ACCEPTED.Ptr(),
		PackageInfo:    core.String(packageInfo),
	}

	ctx.JSON(http.StatusOK, response)
}

type NotifyResp struct {
	ID           string   `json:"id"`
	CreateTime   string   `json:"create_time"`
	ResourceType string   `json:"resource_type"`
	EventType    string   `json:"event_type"`
	Summary      string   `json:"summary"`
	Resource     Resource `json:"resource"`
}

type Resource struct {
	OriginalType   string `json:"original_type"`
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
}

type DecryptResult struct {
	OutBillNo      string `json:"out_bill_no"`
	TransferBillNo string `json:"transfer_bill_no"`
	State          string `json:"state"`
	MchId          string `json:"mch_id"`
	TransferAmount int64  `json:"transfer_amount"`
	Openid         string `json:"openid"`
	FailReason     string `json:"fail_reason"`
	CreateTime     string `json:"create_time"`
	UpdateTime     string `json:"update_time"`
}

func (t *TransferHandler) TransferNotify(ctx *gin.Context) {
	// 1. 构造回调体
	var resp struct {
		OutBillNo string `json:"out_bill_no"  binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&resp); err != nil {
		ctx.JSON(400, gin.H{"code": "FAIL", "message": "invalid body"})
		return
	}

	// 更新	 requestRecord 状态
	err := t.svc.UpdateTransferStatus(ctx, resp.OutBillNo, domain.TransferStatusWaitUserConfirm)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.String(http.StatusOK, "")
}

// 解密 AES-256-GCM 回调
// apiV3Key 必须是 32 字节字符串
func DecryptNotifyResource(apiV3Key, associatedData, nonce, ciphertext string) (string, error) {
	jsonStr := `{
		"out_bill_no": "plfk2020042013",
		"transfer_bill_no":"1330000071100999991182020050700019480001",
		"state": "SUCCESS",
		"mch_id": "1900001109",
		"transfer_amount": 2000,
		"openid": "o-MYE421800elYMDE34nYD456Xoy",
		"fail_reason": "PAYEE_ACCOUNT_ABNORMAL",
		"create_time": "2015-05-20T13:29:35+08:00",
		"update_time": "2023-08-15T20:33:22+08:00"
	}`
	return jsonStr, nil
	key := []byte(apiV3Key)
	if len(key) != 32 {
		return "", errors.New("无效的ApiV3Key，长度必须为32个字节")
	}

	nonceBytes := []byte(nonce)
	aad := []byte(associatedData)
	ct, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plain, err := gcm.Open(nil, nonceBytes, ct, aad)
	if err != nil {
		return "", err
	}
	return string(plain), nil

}

// 判断 notify 是不是来了

func (t *TransferHandler) ConfirmTransfer(ctx *gin.Context) {
	var req struct {
		MchId       string `json:"mch_id" binding:"required"`
		Appid       string `json:"appid" binding:"required"`
		PackageInfo string `json:"package_info" binding:"required"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法: " + err.Error()})
		return
	}

	record, err := t.svc.GetTransferRecordByPackageInfo(ctx, req.PackageInfo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	if record.Status == domain.TransferStatusWaitUserConfirm {
		ctx.String(http.StatusOK, "")
		// 如果状态为 TransferStatusWaitUserConfirm，则更新用户余额

		err := t.userSvc.UpdateBalance(ctx, record.Openid, record.Amount)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "")
			log.Printf("更新用户余额失败: %v", err)
			return
		}
		err = t.svc.UpdateTransferStatus(ctx, record.OutBillNo, domain.TransferStatusSuccess)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, "")
			log.Printf("更新转账状态失败: %v", err)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "转账确认成功"})
	} else {
		ctx.String(http.StatusInternalServerError, "")
	}

}

func (t *TransferHandler) FetchAmount(ctx *gin.Context) {
	openid := ctx.Query("openid")
	log.Println("openid", openid)
	amount, err := t.userSvc.GetAmount(ctx, openid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, 0)
	}
	ctx.JSON(http.StatusOK, amount)
}
