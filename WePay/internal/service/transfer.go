package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"wepay/internal/domain"
	"wepay/internal/repository"
	"wepay/internal/service/wxpay_utility"
)

type TransferService interface {
	TransferToUser(config *wxpay_utility.MchConfig, request *TransferToUserRequest) (response *TransferToUserResponse, err error)
	GenerateOutBillNo(openid string, amount int64) string
	AddTransferRequest(ctx context.Context, req *domain.TransferRecord) error
	GetTransferStatus(ctx context.Context, outbillno string) (string, error)
	UpdateTransferStatus(ctx context.Context, outbillno, state string) error
	GetTransferRecord(ctx context.Context, outbillno string) (domain.TransferRecord, error)
}

type transferService struct {
	repo repository.TransferRepository
}

func NewTransferService(repo repository.TransferRepository) TransferService {
	return &transferService{
		repo: repo,
	}
}

// TransferToUser 发起转账到用户
func (svc *transferService) TransferToUser(config *wxpay_utility.MchConfig, request *TransferToUserRequest) (response *TransferToUserResponse, err error) {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "POST"
		path   = "/v3/fund-app/mch-transfer/transfer-bills"
	)

	reqUrl, err := url.Parse(fmt.Sprintf("%s%s", host, path))
	if err != nil {
		return nil, err
	}
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	httpRequest, err := http.NewRequest(method, reqUrl.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Wechatpay-Serial", config.WechatPayPublicKeyId())
	httpRequest.Header.Set("Content-Type", "application/json")
	authorization, err := wxpay_utility.BuildAuthorization(config.MchId(), config.CertificateSerialNo(), config.PrivateKey(), method, reqUrl.Path, reqBody)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Authorization", authorization)

	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	respBody, err := wxpay_utility.ExtractResponseBody(httpResponse)
	if err != nil {
		return nil, err
	}

	if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300 {
		// 2XX 成功，验证应答签名
		err = wxpay_utility.ValidateResponse(
			config.WechatPayPublicKeyId(),
			config.WechatPayPublicKey(),
			&httpResponse.Header,
			respBody,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(respBody, response); err != nil {
			return nil, err
		}

		return response, nil
	} else {
		return nil, wxpay_utility.NewApiException(
			httpResponse.StatusCode,
			httpResponse.Header,
			respBody,
		)
	}
}

func (svc *transferService) GenerateOutBillNo(openid string, amount int64) string {
	return fmt.Sprintf("Transfer_%v_%v_%v", openid, amount, strconv.FormatInt(time.Now().UnixNano(), 10))
}

func (svc *transferService) AddTransferRequest(ctx context.Context, req *domain.TransferRecord) error {

	err := svc.repo.CreateTransferRequest(ctx, req)
	if err != nil {
		log.Printf("Failed to insert into database for TransferRecord: %v", err)
	}
	return err
}

func (svc *transferService) GetTransferStatus(ctx context.Context, outbillno string) (string, error) {
	return svc.repo.GetTransferStatus(ctx, outbillno)
}

func (svc *transferService) UpdateTransferStatus(ctx context.Context, outbillno, state string) error {
	return svc.repo.UpdateTransferRequestStatus(ctx, outbillno, state)
}

func (svc *transferService) GetTransferRecord(ctx context.Context, outbillno string) (domain.TransferRecord, error) {
	return svc.repo.GetTransferRecord(ctx, outbillno)
}
