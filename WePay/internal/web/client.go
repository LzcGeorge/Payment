package web

import "wepay/internal/service/wxpay_utility"

type Client struct {
	Appid     string
	MchConfig *wxpay_utility.MchConfig
	NotifyUrl string
}

func NewClient(appid string, mchConfig *wxpay_utility.MchConfig, notifyUrl string) Client {
	return Client{
		Appid:     appid,
		MchConfig: mchConfig,
		NotifyUrl: notifyUrl,
	}
}
