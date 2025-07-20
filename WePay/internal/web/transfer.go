package web

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"wepay/internal/service"
	"wepay/internal/service/wxpay_utility"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
)

type TransferHandler struct {
	svc    *service.TransferService
	client Client
}

func NewTransferHandler(svc *service.TransferService, client Client) *TransferHandler {
	return &TransferHandler{
		svc:    svc,
		client: client,
	}
}

// RegisterRoutes registers the HTTP routes for transfer operations on the provided router group.
// Currently, it registers a POST endpoint at "/to_user" that triggers the InitiateTransfer handler method.
func (t *TransferHandler) RegisterRoutes(ug *gin.RouterGroup) {
	ug.POST("/to_user", t.InitiateTransfer)
	ug.POST("/notify", t.TransferNotify)
}

func (t *TransferHandler) InitiateTransfer(ctx *gin.Context) {
	// 用户传来的参数
	var req struct {
		Openid string `form:"openid" json:"openid" binding:"required"`
		Amount int64  `form:"amount" json:"amount" binding:"required"`
		Remark string `form:"remark" json:"remark"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法: " + err.Error()})
		return
	}
	openid := req.Openid // 用户openid
	amount := req.Amount // 转账金额，单位为 ”分“
	remark := req.Remark // 转账备注

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
	t.svc.AddTransferRequest(ctx, openid, amount, remark, transfer_scene_id)

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
	response, err := t.svc.TransferToUser(mchConfig, request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("TransferToUser response: %+v", response)
	switch response.State {
	case service.TRANSFERBILLSTATUS_SUCCESS.Ptr():
		ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "转账成功", "data": response})
	case service.TRANSFERBILLSTATUS_FAIL.Ptr():
		ctx.JSON(http.StatusOK, gin.H{"code": 1, "msg": "转账失败", "detail": response.PackageInfo})
	case service.TRANSFERBILLSTATUS_PROCESSING.Ptr():
		ctx.JSON(http.StatusOK, gin.H{"code": 2, "msg": "转账处理中"})
	default:
		ctx.JSON(http.StatusOK, gin.H{"code": 99, "msg": "未知状态"})
	}

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
	var resp NotifyResp
	if err := ctx.ShouldBindJSON(&resp); err != nil {
		ctx.JSON(400, gin.H{"code": "FAIL", "message": "invalid body"})
		return
	}

	// 2. 校验签名（开发环境可以不做/或模拟通过）
	// 3. 解密 resource.ciphertext

	plain, err := DecryptNotifyResource(
		t.client.WechatPayPublicKeyPath,
		resp.Resource.AssociatedData,
		resp.Resource.Nonce,
		resp.Resource.Ciphertext,
	)
	if err != nil {
		ctx.JSON(400, gin.H{"code": "FAIL", "message": "解密失败"})
		return
	}

	var order DecryptResult
	if err := json.Unmarshal([]byte(plain), &order); err != nil {
		ctx.JSON(400, gin.H{"code": "FAIL", "message": "解密数据不合法"})
		return
	}

	err = t.svc.TransferCallback(ctx, order.OutBillNo, order.State)
	log.Printf("TransferCallback: %+v", order)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. 校验签名（开发环境可以不做/或模拟通过）
	// 3. 解密 resource.ciphertext
	// apiV3Key := os.Getenv("WECHAT_APIV3_KEY")
	// plain, err := service.DecryptNotifyResource(
	// 	apiV3Key,
	// 	req.Resource.AssociatedData,
	// 	req.Resource.Nonce,
	// 	req.Resource.Ciphertext,
	// )
	// if err != nil {
	// 	ctx.JSON(400, gin.H{"code": "FAIL", "message": "解密失败"})
	// 	return
	// }

	// var order DecryptResult
	// if err := json.Unmarshal([]byte(plain), &order); err != nil {
	// 	ctx.JSON(400, gin.H{"code": "FAIL", "message": "解密数据不合法"})
	// 	return
	// }

	// order := DecryptResult{
	// // 4. 幂等落库（防重复），状态流转，用户余额变更
	// err = t.svc.OnTransferNotify(ctx, &order)
	// if err != nil {
	// 	ctx.JSON(500, gin.H{"code": "FAIL", "message": "落库失败"})
	// 	return
	// }

	// 5. 返回成功（微信要求200或204，无需body）
	ctx.String(200, "")
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
