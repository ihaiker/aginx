package admin

import (
	"bytes"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/hero"
	"github.com/rs/cors"
	"io/ioutil"
	"net/http"
	"strings"
)

type Node struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func Routers(adminConfig string) func(app *iris.Application) {
	nodes := NewNodeController(adminConfig)

	//跨域处理
	return func(app *iris.Application) {
		//静态文件引入
		Static(app)

		app.WrapRouter(cors.New(cors.Options{
			AllowOriginFunc:  func(origin string) bool { return true },
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders:   []string{"*"},
		}).ServeHTTP)

		h := hero.New()
		admin := app.Party("/admin")
		{
			admin.Get("/nodes", h.Handler(func(ctx context.Context) []*Node {
				return nodes.list(true)
			}))
			admin.Post("/node", h.Handler(nodes.add))
			admin.Delete("/node", h.Handler(nodes.delete))

			admin.Any("**", func(ctx context.Context) {
				nodeName := ctx.GetHeader("Aginxnode")
				node := nodes.getNode(nodeName)
				errors.Assert(node != nil, "未发现节点%s", nodeName)

				reqUrl := node.Address + ctx.Request().RequestURI[6:]
				var req *http.Request
				if body, err := ctx.GetBody(); err == nil {
					req, err = http.NewRequest(ctx.Method(), reqUrl, bytes.NewBuffer(body))
				} else {
					req, err = http.NewRequest(ctx.Method(), reqUrl, nil)
				}
				for _, cookie := range ctx.Request().Cookies() {
					req.AddCookie(cookie)
				}
				req.Header = ctx.Request().Header
				req.SetBasicAuth(node.User, node.Password)

				resp, err := http.DefaultClient.Do(req)
				errors.PanicMessage(err, "代理异常")

				ctx.StatusCode(resp.StatusCode)
				for name, values := range resp.Header {
					ctx.ResponseWriter().Header().Set(name, strings.Join(values, " "))
				}

				out, err := ioutil.ReadAll(resp.Body)
				errors.PanicMessage(err, "代理异常返回")
				_, _ = ctx.Write(out)
			})
		}
	}
}
