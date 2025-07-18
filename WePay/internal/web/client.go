package web

type Client struct {
	Appid                  string
	Mchid                  string
	CertificateSerialNo    string
	PrivateKeyPath         string
	WechatPayPublicKeyId   string
	WechatPayPublicKeyPath string
}

func NewClient(appid, mchid, certificateSerialNo, privateKeyPath, wechatPayPublicKeyId, wechatPayPublicKeyPath string) *Client {
	return &Client{
		Appid:                  appid,
		Mchid:                  mchid,
		CertificateSerialNo:    certificateSerialNo,
		PrivateKeyPath:         privateKeyPath,
		WechatPayPublicKeyId:   wechatPayPublicKeyId,
		WechatPayPublicKeyPath: wechatPayPublicKeyPath,
	}
}
