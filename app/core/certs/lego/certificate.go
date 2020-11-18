package lego

import (
	"time"
)

//证书文件
type Certificate struct {
	Email         string    `json:"email"`
	Domain        string    `json:"domain"`
	CertURL       string    `json:"certUrl"`
	CertStableURL string    `json:"certStableUrl"`
	ExpireTime    time.Time `json:"expireTime"`

	CertificatePath string `json:"certificatePath"`
	PrivateKeyPath  string `json:"privateKeyPath"`
}

func (c *Certificate) GetExpiredTime() time.Time {
	return c.ExpireTime
}

func (c *Certificate) Certificate() string {
	return c.CertificatePath
}

func (c *Certificate) PrivateKey() string {
	return c.PrivateKeyPath
}
