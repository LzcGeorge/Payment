package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
	"wepay/internal/domain"
	"wepay/internal/service"
	"wepay/internal/service/wxpay_utility"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
)

type TransferHandler struct {
	svc          service.TransferService
	userSvc      service.UserService
	client       Client
	notifyChans  sync.Map // map[string]chan struct{}
	confirmChans sync.Map // map[string]chan struct{}
}

func NewTransferHandler(svc service.TransferService, userSvc service.UserService, client Client) *TransferHandler {
	return &TransferHandler{
		svc:          svc,
		userSvc:      userSvc,
		client:       client,
		notifyChans:  sync.Map{},
		confirmChans: sync.Map{},
	}
}

func (t *TransferHandler) RegisterRoutes(ug *gin.RouterGroup) {
	ug.POST("/to_user", t.InitiateTransfer)
	ug.POST("/notify", t.TransferNotify)   // 微信支付的回调（手动模拟实现）
	ug.POST("/confirm", t.ConfirmTransfer) // 确认转账
	ug.GET("/amount", t.FetchAmount)       // 查询余额
}

func generatePackageInfo(openid string, timeStr string) string {
	return fmt.Sprintf("PK%s-%s", openid, timeStr)
}

// 发起转账
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

	// 商户的配置 transfer_scene_id（转账场景）,  user_recv_perception（用户收款码）
	transfer_scene_id := "1000"    // 转账场景：现金营销
	user_recv_perception := "现金红包" // 用户收款时感知到的收款原因将根据转账场景自动展示默认内容。

	// 生成唯一outbillno, packageInfo并保存转账请求
	outbillno := t.svc.GenerateOutBillNo(req.Openid, req.Amount)
	packageInfo := generatePackageInfo(req.Openid, req.Time)
	t.userSvc.UpsertUser(ctx, req.Openid)
	requestRecord := &domain.TransferRecord{
		OutBillNo:   outbillno,
		Openid:      req.Openid,
		MchId:       t.client.MchConfig.MchId(),
		PackageInfo: packageInfo,
		Amount:      req.Amount,
		Remark:      req.Remark,
		Status:      domain.TransferStatusProcessing,
	}
	log.Println("outbillno", outbillno)
	err := t.svc.AddTransferRequest(ctx, requestRecord)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Println("add transfer request error:", err)
		return
	}

	// 构造 TransferToUserRequest
	request := &service.TransferToUserRequest{
		// 商家
		Appid:              core.String(t.client.Appid), // 小程序与商户关联的appid
		OutBillNo:          core.String(outbillno),
		TransferSceneId:    core.String(transfer_scene_id),
		Openid:             core.String(req.Openid),
		MchId:              core.String(t.client.MchConfig.MchId()),
		UserName:           core.String(req.Openid),
		TransferAmount:     core.Int64(req.Amount),
		TransferRemark:     core.String(req.Remark),
		NotifyUrl:          core.String(t.client.NotifyUrl),
		UserRecvPerception: core.String(user_recv_perception),
	}

	// 发起转账
	_, err = t.svc.TransferToUser(t.client.MchConfig, request)
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

	go func() {
		time.Sleep(5 * time.Second)
		t.svc.UpdateTransferStatus(ctx, outbillno, domain.TransferStatusWaitUserConfirm)
	}()

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
	// 1. 请求体
	var req struct {
		OutBillNo string `json:"out_bill_no"  binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"code": "FAIL", "message": "invalid body"})
		return
	}

	// 2. 校验回调请求
	headers := ctx.Request.Header
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	err = wxpay_utility.ValidateResponse(t.client.MchConfig.WechatPayPublicKeyId(), t.client.MchConfig.WechatPayPublicKey(), &headers, body)
	if err != nil {
		// ctx.JSON(http.StatusInternalServerError, err.Error())
		log.Println("validate response error:", err)
	}

	// 是否有确认转账的请求
	_, ok := t.confirmChans.Load(req.OutBillNo)
	if !ok {
		log.Println("wait confirm")
		ctx.JSON(http.StatusOK, gin.H{"message": "wait confirm"})
		return
	}

	// 唤醒 notify 的协程
	defer func() {
		t.notifyChans.Store(req.OutBillNo, make(chan struct{}))
		log.Println("notify 唤醒")
	}()

	ctx.String(http.StatusOK, "")

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

	// 获取转账记录
	record, err := t.svc.GetTransferRecordByPackageInfo(ctx, req.PackageInfo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	// 唤醒 confirm 的协程
	t.confirmChans.Store(record.OutBillNo, make(chan struct{}))
	log.Println("confirm 来了")

	// 等待 notify 传来消息
	timeout := time.After(10 * time.Second)
	interval := 1 * time.Second
	for {
		select {
		case <-timeout:
			ctx.JSON(http.StatusRequestTimeout, gin.H{"message": "超时未收到回调"})
			log.Println("超时未收到回调")
			return
		default:
			log.Println("wait notify")
			ch, ok := t.notifyChans.Load(record.OutBillNo)
			if ok {
				close(ch.(chan struct{}))
				t.notifyChans.Delete(record.OutBillNo)
				// notify 来了
				log.Println("notify 来了")
				if record.Status == domain.TransferStatusWaitUserConfirm {
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
					return
				} else {
					ctx.JSON(http.StatusInternalServerError, "转账状态不正确")
					log.Println("转账状态不正确")
					return
				}
			} else {
				time.Sleep(interval)
			}
		}
	}

}

func (t *TransferHandler) FetchAmount(ctx *gin.Context) {
	openid := ctx.Query("openid")
	log.Println("openid", openid)
	amount, err := t.userSvc.GetAmount(ctx, openid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, 0)
		return
	}
	ctx.JSON(http.StatusOK, amount)
}
