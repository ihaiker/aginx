package auth

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"github.com/ihaiker/aginx/v2/core/config"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"gopkg.in/ldap.v2"
	"io/ioutil"
	"text/template"
)

func filter(filterTemp, attr, user string) string {
	t, err := template.New("").Parse(filterTemp)
	errors.PanicMessage(err, "查询filter文件错误")
	out := bytes.NewBufferString("")
	err = t.Execute(out, map[string]string{
		"UsernameAttribute": attr,
		"Username":          user,
	})
	errors.PanicMessage(err, "查询filter文件错误")
	return out.String()
}

func check(username, password string) bool {
	if config.Config.Auth.Mode == "basic" {
		pwd, has := config.Config.Auth.Users[username]
		errors.Assert(has, "用户不存在！")
		errors.Assert(password == pwd, "用户名或者密码错误！")
	} else {
		var l *ldap.Conn
		var err error
		if config.Config.Auth.LDAP.TLSCa == "" {
			l, err = ldap.Dial("tcp", config.Config.Auth.LDAP.Server)
		} else {
			// Load client cert and key
			cert, err := tls.LoadX509KeyPair(config.Config.Auth.LDAP.TLSCert, config.Config.Auth.LDAP.TLSKey)
			errors.PanicMessage(err, "加载证书位置")
			// Load CA chain
			caCert, err := ioutil.ReadFile(config.Config.Auth.LDAP.TLSCa)
			errors.PanicMessage(err, "加载证书位置")
			//
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      caCertPool, InsecureSkipVerify: true,
			}
			l, err = ldap.DialTLS("tcp", config.Config.Auth.LDAP.Server, tlsConfig)
		}
		errors.PanicMessage(err, "连接LDAP服务")

		// 绑定用于管理的用户
		err = l.Bind(config.Config.Auth.LDAP.BindDn, config.Config.Auth.LDAP.Password)
		errors.PanicMessage(err, "连接LDAP密码错误")

		// 查询
		sql := ldap.NewSearchRequest(config.Config.Auth.LDAP.BaseDn,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			filter(config.Config.Auth.LDAP.Filter, config.Config.Auth.LDAP.UsernameAttribute, username),
			nil, nil)

		cur, err := l.Search(sql)
		errors.PanicMessage(err, "搜索LDAP用户错误")
		errors.Assert(len(cur.Entries) == 1, "获取LDAP用户信息错误")

		err = l.Bind(cur.Entries[0].DN, password)
		errors.Assert(len(cur.Entries) == 1, "验证LDAP用户信息错误")
		errors.PanicMessage(err, "用户LDAP错误")
	}
	return true
}
