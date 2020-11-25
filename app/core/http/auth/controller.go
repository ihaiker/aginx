package auth

import (
	"encoding/base64"
	"fmt"
	"github.com/ihaiker/aginx/v2/core/config"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
)

func Login(authCfg config.Auth) interface{} {
	if authCfg.LDAP.BaseDn == "" {
		authCfg.LDAP.BaseDn = authCfg.LDAP.BindDn
	}

	return func(ctx iris.Context) string {
		auth := map[string]string{}
		_ = ctx.ReadJSON(&auth)
		username := auth["username"]
		password := auth["password"]
		errors.Assert(username != "" && password != "", "用户名或者密码不能为空！")

		check(username, password)

		fullUser := fmt.Sprintf("%s:%s", username, password)
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(fullUser))
	}
}
