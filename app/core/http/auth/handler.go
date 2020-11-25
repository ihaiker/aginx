package auth

import (
	"github.com/ihaiker/aginx/v2/core/config"
	"github.com/ihaiker/aginx/v2/core/logs"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
)

func Handler(authCfg config.Auth) iris.Handler {
	if authCfg.LDAP.BaseDn == "" {
		authCfg.LDAP.BaseDn = authCfg.LDAP.BindDn
	}

	return func(ctx iris.Context) {
		user, password, ok := ctx.Request().BasicAuth()
		if !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			return
		}
		err := errors.Safe(func() {
			check(user, password)
		})
		if err != nil {
			logs.Debug("认证错误：user:", user, ", password:", password, ", err:", err)
			ctx.StatusCode(iris.StatusUnauthorized)
		} else {
			ctx.Next()
		}
	}
}
