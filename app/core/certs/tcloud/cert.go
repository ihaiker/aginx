package tcloud

import "time"

type certStatus string

const (
	Normal     certStatus = "normal"     //正常状态
	NoVerified certStatus = "noVerified" //待验证
	Expired    certStatus = "expired"    //过期
)

type tencentCertificate struct {
	certId      string     //证书ID
	domain      string     //域名
	status      certStatus //状态
	expiredTime time.Time
	insertTime  time.Time

	loc string //认证路径
	txt string //认证字符

	certificate string //证书内容
	privateKey  string //证书私钥

	certificatePath string //证书路径
	privateKeyPath  string //证书私钥路径
}

func (t *tencentCertificate) GetExpiredTime() time.Time {
	return t.expiredTime
}

func (t *tencentCertificate) Certificate() string {
	return t.certificatePath
}

func (t *tencentCertificate) PrivateKey() string {
	return t.privateKeyPath
}
