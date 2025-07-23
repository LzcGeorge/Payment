package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"wepay/internal/service"
	svcmocks "wepay/internal/service/mocks"
	"wepay/internal/service/wxpay_utility"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"go.uber.org/mock/gomock"
)

func TestInitiateTransfer(t *testing.T) {
	testCases := []struct {
		name     string
		reqBody  string
		mock     func(ctrl *gomock.Controller) service.TransferService
		wantCode int
		wantResp service.TransferToUserResponse
	}{
		{
			name: "success",
			reqBody: `{
				"openid": "o1234567890",
				"amount": 100,
				"remark": "test",
				"time": "20200420130000"
			}`,
			mock: func(ctrl *gomock.Controller) service.TransferService {
				transferSvc := svcmocks.NewMockTransferService(ctrl)
				transferSvc.EXPECT().GenerateOutBillNo(gomock.Any(), gomock.Any()).Return("plfk2020042013")
				transferSvc.EXPECT().AddTransferRequest(gomock.Any(), gomock.Any()).Return(nil)

				transferSvc.EXPECT().TransferToUser(gomock.Any(), gomock.Any()).Return(&service.TransferToUserResponse{
					OutBillNo:      core.String("plfk2020042013"),
					TransferBillNo: core.String("1330000071100999991182020050700019480001"),
					CreateTime:     core.String("2015-05-20T13:29:35.120+08:00"),
					State:          service.TRANSFERBILLSTATUS_WAIT_USER_CONFIRM.Ptr(),
					PackageInfo:    core.String("PKo1234567890-20200420130000"),
				}, nil)

				return transferSvc
			},
			wantCode: http.StatusOK,
			wantResp: service.TransferToUserResponse{
				OutBillNo:      core.String("plfk2020042013"),
				TransferBillNo: core.String("1330000071100999991182020050700019480001"),
				CreateTime:     core.String("2015-05-20T13:29:35.120+08:00"),
				State:          service.TRANSFERBILLSTATUS_WAIT_USER_CONFIRM.Ptr(),
				PackageInfo:    core.String("PKo1234567890-20200420130000"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 创建 userHandler 及所需的依赖 userService
			server := gin.Default()
			transferSvc := tc.mock(ctrl)
			MchConfig, _ := wxpay_utility.CreateMchConfig(
				"1368139500",
				"ajkhyuiKJSAHDn124fsadasda",
				"certs/private_key.pem",
				"adsbvcretgnfsde",
				"certs/public_key.pem",
			)
			client := NewClient("wxb9f4f763e5d4a6de", MchConfig, "http://wepay.selfknow.cn")
			transferHandler := NewTransferHandler(transferSvc, nil, client)
			transferHandler.RegisterRoutes(server.Group("/transfer"))

			// 创建请求
			req, err := http.NewRequest(http.MethodPost, "/transfer/to_user", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			assert.Nil(t, err)

			// 执行请求
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			// 检查响应
			var respBody service.TransferToUserResponse
			err = json.Unmarshal(resp.Body.Bytes(), &respBody)
			assert.Nil(t, err)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantResp, respBody)
		})
	}
}
