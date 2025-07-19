package web

import (
	"net/http"
	"os"
	"wepay/internal/service"
	"wepay/internal/service/wxpay_utility"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
)

type TransferHandler struct {
	svc *service.TransferService
}

func NewTransferHandler(svc *service.TransferService) *TransferHandler {
	return &TransferHandler{
		svc: svc,
	}
}

func (t *TransferHandler) RegisterRoutes(ug *gin.RouterGroup) {
	ug.POST("/to_user", t.InitiateTransfer)
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

	notify_url := os.Getenv("WECHAT_NOTIFY_URL")
	user_recv_perception := "现金红包" // 用户收款时感知到的收款原因将根据转账场景自动展示默认内容。

	mchConfig, appid, err := t.getMchConfigAndAppid()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一outbillno并保存转账请求
	outbillno := t.svc.GenerateOutBillNo(openid, amount)

	// 构造 TransferToUserRequest
	request := &service.TransferToUserRequest{
		// 商家
		Appid:              core.String(appid), // 小程序与商户关联的appid
		OutBillNo:          core.String(outbillno),
		TransferSceneId:    core.String(transfer_scene_id),
		Openid:             core.String(openid),
		UserName:           core.String(remark),
		TransferAmount:     core.Int64(amount),
		TransferRemark:     core.String(remark),
		NotifyUrl:          core.String(notify_url),
		UserRecvPerception: core.String(user_recv_perception),
	}
	// func (svc *TransferService) TransferToUser(config *wxpay_utility.MchConfig, request *TransferToUserRequest) (response *TransferToUserResponse, err error) {
	response, err := t.svc.TransferToUser(mchConfig, request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch response.State {
	case service.TRANSFERBILLSTATUS_SUCCESS.Ptr():
		ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "转账成功", "data": response})
		t.svc.AddTransferRequest(ctx, openid, amount, remark, transfer_scene_id)
	case service.TRANSFERBILLSTATUS_FAIL.Ptr():
		ctx.JSON(http.StatusOK, gin.H{"code": 1, "msg": "转账失败", "detail": response.PackageInfo})
	case service.TRANSFERBILLSTATUS_PROCESSING.Ptr():
		ctx.JSON(http.StatusOK, gin.H{"code": 2, "msg": "转账处理中"})
	default:
		ctx.JSON(http.StatusOK, gin.H{"code": 99, "msg": "未知状态"})
	}

}

func (t *TransferHandler) getMchConfigAndAppid() (mch *wxpay_utility.MchConfig, appid string, err error) {

	appid = "test_appid"
	mchId := "test_mch_id"
	certificateSerialNo := "test_serial_no"
	privateKeyPath := "./certs/test_private_key.pem"
	wechatPayPublicKeyId := "test_public_key_id"
	wechatPayPublicKeyPath := "./certs/test_public_key.pem"

	// 从环境变量获取配置，方便测试时切换

	mchId = os.Getenv("WECHAT_MCH_ID")
	if mchId != "" {
		certificateSerialNo = os.Getenv("WECHAT_CERT_SERIAL_NO")
		privateKeyPath = os.Getenv("WECHAT_PRIVATE_KEY_PATH")
		wechatPayPublicKeyId = os.Getenv("WECHAT_PUBLIC_KEY_ID")
		wechatPayPublicKeyPath = os.Getenv("WECHAT_PUBLIC_KEY_PATH")
	}
	mchId = "test_mch_id" // 测试用的商户号
	mch, err = wxpay_utility.CreateMchConfig(
		mchId,
		certificateSerialNo,
		privateKeyPath,
		wechatPayPublicKeyId,
		wechatPayPublicKeyPath,
	)

	return mch, appid, err
}
