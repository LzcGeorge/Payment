package wxpay_utility

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"
)

type MchConfigInterface interface {
	MchId() string
	CertificateSerialNo() string
	PrivateKey() *rsa.PrivateKey
	WechatPayPublicKeyId() string
	WechatPayPublicKey() *rsa.PublicKey
}

// MchConfig 商户信息配置，用于调用商户API
type MchConfig struct {
	mchId                      string
	certificateSerialNo        string
	privateKeyFilePath         string
	wechatPayPublicKeyId       string
	wechatPayPublicKeyFilePath string
	privateKey                 *rsa.PrivateKey
	wechatPayPublicKey         *rsa.PublicKey
}

// MchId 商户号
func (c *MchConfig) MchId() string {
	return c.mchId
}

// CertificateSerialNo 商户API证书序列号
func (c *MchConfig) CertificateSerialNo() string {
	return c.certificateSerialNo
}

// PrivateKey 商户API证书对应的私钥
func (c *MchConfig) PrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

// WechatPayPublicKeyId 微信支付公钥ID
func (c *MchConfig) WechatPayPublicKeyId() string {
	return c.wechatPayPublicKeyId
}

// WechatPayPublicKey 微信支付公钥
func (c *MchConfig) WechatPayPublicKey() *rsa.PublicKey {
	return c.wechatPayPublicKey
}

// CreateMchConfig MchConfig 构造函数
func CreateMchConfig(
	mchId string,
	certificateSerialNo string,
	privateKeyFilePath string,
	wechatPayPublicKeyId string,
	wechatPayPublicKeyFilePath string,
) (*MchConfig, error) {
	mchConfig := &MchConfig{
		mchId:                      mchId,
		certificateSerialNo:        certificateSerialNo,
		privateKeyFilePath:         privateKeyFilePath,
		wechatPayPublicKeyId:       wechatPayPublicKeyId,
		wechatPayPublicKeyFilePath: wechatPayPublicKeyFilePath,
	}
	privateKey, err := LoadPrivateKeyWithPath(mchConfig.privateKeyFilePath)
	if err != nil {
		//return nil, err
		privateKey = &rsa.PrivateKey{
			PublicKey: rsa.PublicKey{
				N: big.NewInt(0), // 公钥N，实际应为大素数乘积
				E: 65537,         // 公钥e，常用65537
			},
			D:      big.NewInt(0),                            // 私钥d，实际应为大整数
			Primes: []*big.Int{big.NewInt(0), big.NewInt(0)}, // 两个大素数p、q
		}
		log.Printf("load private key from file %s error: %s", mchConfig.privateKeyFilePath, err.Error())
	}
	mchConfig.privateKey = privateKey
	wechatPayPublicKey, err := LoadPublicKeyWithPath(mchConfig.wechatPayPublicKeyFilePath)
	if err != nil {
		//return nil, err
		wechatPayPublicKey = &rsa.PublicKey{
			N: big.NewInt(0), // 公钥N，实际应为大素数乘积
			E: 65537,         // 公钥e，常用65537
		}
		log.Printf("load wechat pay public key from file %s error: %s", mchConfig.wechatPayPublicKeyFilePath, err.Error())
	}
	mchConfig.wechatPayPublicKey = wechatPayPublicKey
	return mchConfig, nil
}

// LoadPrivateKey 通过私钥的文本内容加载私钥
func LoadPrivateKey(privateKeyStr string) (privateKey *rsa.PrivateKey, err error) {
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return nil, fmt.Errorf("decode private key err")
	}
	if block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("the kind of PEM should be PRVATE KEY")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key err:%s", err.Error())
	}
	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not a RSA private key")
	}
	return privateKey, nil
}

// LoadPublicKey 通过公钥的文本内容加载公钥
func LoadPublicKey(publicKeyStr string) (publicKey *rsa.PublicKey, err error) {
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil {
		return nil, errors.New("decode public key error")
	}
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("the kind of PEM should be PUBLIC KEY")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key err:%s", err.Error())
	}
	publicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%s is not rsa public key", publicKeyStr)
	}
	return publicKey, nil
}

// LoadPrivateKeyWithPath 通过私钥的文件路径内容加载私钥
func LoadPrivateKeyWithPath(path string) (privateKey *rsa.PrivateKey, err error) {
	privateKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private pem file err:%s", err.Error())
	}
	return LoadPrivateKey(string(privateKeyBytes))
}

// LoadPublicKeyWithPath 通过公钥的文件路径加载公钥
func LoadPublicKeyWithPath(path string) (publicKey *rsa.PublicKey, err error) {
	publicKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read certificate pem file err:%s", err.Error())
	}
	return LoadPublicKey(string(publicKeyBytes))
}

// EncryptOAEPWithPublicKey 使用 OAEP padding方式用公钥进行加密
func EncryptOAEPWithPublicKey(message string, publicKey *rsa.PublicKey) (ciphertext string, err error) {
	if publicKey == nil {
		return "", fmt.Errorf("you should input *rsa.PublicKey")
	}
	ciphertextByte, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, publicKey, []byte(message), nil)
	if err != nil {
		return "", fmt.Errorf("encrypt message with public key err:%s", err.Error())
	}
	ciphertext = base64.StdEncoding.EncodeToString(ciphertextByte)
	return ciphertext, nil
}

// SignSHA256WithRSA 通过私钥对字符串以 SHA256WithRSA 算法生成签名信息
func SignSHA256WithRSA(source string, privateKey *rsa.PrivateKey) (signature string, err error) {
	if privateKey == nil {
		return "", fmt.Errorf("private key should not be nil")
	}
	h := crypto.Hash.New(crypto.SHA256)
	_, err = h.Write([]byte(source))
	if err != nil {
		return "", nil
	}
	hashed := h.Sum(nil)
	signatureByte, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signatureByte), nil
}

// VerifySHA256WithRSA 通过公钥对字符串和签名结果以 SHA256WithRSA 验证签名有效性
func VerifySHA256WithRSA(source string, signature string, publicKey *rsa.PublicKey) error {
	if publicKey == nil {
		return fmt.Errorf("public key should not be nil")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("verify failed: signature is not base64 encoded")
	}
	hashed := sha256.Sum256([]byte(source))
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], sigBytes)
	if err != nil {
		return fmt.Errorf("verify signature with public key error:%s", err.Error())
	}
	return nil
}

// GenerateNonce 生成一个长度为 NonceLength 的随机字符串（只包含大小写字母与数字）
func GenerateNonce() (string, error) {
	const (
		// NonceSymbols 随机字符串可用字符集
		NonceSymbols = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		// NonceLength 随机字符串的长度
		NonceLength = 32
	)

	bytes := make([]byte, NonceLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	symbolsByteLength := byte(len(NonceSymbols))
	for i, b := range bytes {
		bytes[i] = NonceSymbols[b%symbolsByteLength]
	}
	return string(bytes), nil
}

// BuildAuthorization 构建请求头中的 Authorization 信息
func BuildAuthorization(
	mchid string,
	certificateSerialNo string,
	privateKey *rsa.PrivateKey,
	method string,
	canonicalURL string,
	body []byte,
) (string, error) {
	const (
		SignatureMessageFormat = "%s\n%s\n%d\n%s\n%s\n" // 数字签名原文格式
		// HeaderAuthorizationFormat 请求头中的 Authorization 拼接格式
		HeaderAuthorizationFormat = "WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",timestamp=\"%d\",serial_no=\"%s\",signature=\"%s\""
	)

	nonce, err := GenerateNonce()
	if err != nil {
		return "", err
	}
	timestamp := time.Now().Unix()
	message := fmt.Sprintf(SignatureMessageFormat, method, canonicalURL, timestamp, nonce, body)
	signature, err := SignSHA256WithRSA(message, privateKey)
	if err != nil {
		return "", err
	}
	authorization := fmt.Sprintf(
		HeaderAuthorizationFormat,
		mchid, nonce, timestamp, certificateSerialNo, signature,
	)
	return authorization, nil
}

// ExtractResponseBody 提取应答报文的 Body
func ExtractResponseBody(response *http.Response) ([]byte, error) {
	if response.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body err:[%s]", err.Error())
	}
	response.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

const (
	WechatPayTimestamp = "Wechatpay-Timestamp" // 微信支付回包时间戳
	WechatPayNonce     = "Wechatpay-Nonce"     // 微信支付回包随机字符串
	WechatPaySignature = "Wechatpay-Signature" // 微信支付回包签名信息
	WechatPaySerial    = "Wechatpay-Serial"    // 微信支付回包平台序列号
	RequestID          = "Request-Id"          // 微信支付回包请求ID
)

// ValidateResponse 验证微信支付回包的签名信息
func ValidateResponse(
	wechatpayPublicKeyId string,
	wechatpayPublicKey *rsa.PublicKey,
	headers *http.Header,
	body []byte,
) error {
	requestID := headers.Get(RequestID)
	timestampStr := headers.Get(WechatPayTimestamp)
	serialNo := headers.Get(WechatPaySerial)
	signature := headers.Get(WechatPaySignature)
	nonce := headers.Get(WechatPayNonce)

	// 拒绝过期请求
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %v", err)
	}
	if time.Now().Sub(time.Unix(timestamp, 0)) > 5*time.Minute {
		return errors.New("invalid timestamp")
	}

	if serialNo != wechatpayPublicKeyId {
		return fmt.Errorf(
			"serial-no mismatch: got %s, expected %s, request-id: %s",
			serialNo,
			wechatpayPublicKeyId,
			requestID,
		)
	}

	message := fmt.Sprintf("%s\n%s\n%s\n", timestampStr, nonce, body)
	if err := VerifySHA256WithRSA(message, signature, wechatpayPublicKey); err != nil {
		return fmt.Errorf("invalid signature: %v, request-id: %s", err, requestID)
	}

	return nil
}

// ApiException 微信支付API错误异常，发送HTTP请求成功，但返回状态码不是 2XX 时抛出本异常
type ApiException struct {
	statusCode   int         // 应答报文的 HTTP 状态码
	header       http.Header // 应答报文的 Header 信息
	body         []byte      // 应答报文的 Body 原文
	errorCode    string      // 微信支付回包的错误码
	errorMessage string      // 微信支付回包的错误信息
}

func (c *ApiException) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("api error:[StatusCode: %d, Body: %s", c.statusCode, string(c.body)))
	if len(c.header) > 0 {
		buf.WriteString(" Header: ")
		for key, value := range c.header {
			buf.WriteString(fmt.Sprintf("\n - %v=%v", key, value))
		}
		buf.WriteString("\n")
	}
	buf.WriteString("]")
	return buf.String()
}

func (c *ApiException) StatusCode() int {
	return c.statusCode
}

func (c *ApiException) Header() http.Header {
	return c.header
}

func (c *ApiException) Body() []byte {
	return c.body
}

func (c *ApiException) ErrorCode() string {
	return c.errorCode
}

func (c *ApiException) ErrorMessage() string {
	return c.errorMessage
}

func NewApiException(statusCode int, header http.Header, body []byte) error {
	ret := &ApiException{
		statusCode: statusCode,
		header:     header,
		body:       body,
	}

	bodyObject := map[string]interface{}{}
	if err := json.Unmarshal(body, &bodyObject); err == nil {
		if val, ok := bodyObject["code"]; ok {
			ret.errorCode = val.(string)
		}
		if val, ok := bodyObject["message"]; ok {
			ret.errorMessage = val.(string)
		}
	}

	return ret
}

// Time 复制 time.Time 对象，并返回复制体的指针
func Time(t time.Time) *time.Time {
	return &t
}

// String 复制 string 对象，并返回复制体的指针
func String(s string) *string {
	return &s
}

// Bool 复制 bool 对象，并返回复制体的指针
func Bool(b bool) *bool {
	return &b
}

// Float64 复制 float64 对象，并返回复制体的指针
func Float64(f float64) *float64 {
	return &f
}

// Float32 复制 float32 对象，并返回复制体的指针
func Float32(f float32) *float32 {
	return &f
}

// Int64 复制 int64 对象，并返回复制体的指针
func Int64(i int64) *int64 {
	return &i
}

// Int32 复制 int64 对象，并返回复制体的指针
func Int32(i int32) *int32 {
	return &i
}
