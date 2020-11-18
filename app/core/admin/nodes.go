package admin

import (
	"fmt"
	"github.com/ihaiker/aginx/v2/core/config"
	"github.com/ihaiker/aginx/v2/core/logs"
	nginxConfig "github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/nginx/query"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"github.com/ihaiker/aginx/v2/core/util/files"
	"github.com/kataras/iris/v12"
	"io/ioutil"
)

type nodeController struct {
	path string
}

func NewNodeController(path string) *nodeController {
	nc := &nodeController{path: path}
	return nc
}

func (nc *nodeController) list() []*Node {
	nodes := make([]*Node, 0)

	if files.Exists(nc.path) {
		conf, err := nginxConfig.Parse(nc.path, &nginxConfig.Options{
			Delimiter: true, RemoveBrackets: true, RemoveAnnotation: true,
		})
		if err != nil {
			logs.Error("解析文件 ", err.Error())
		} else {
			for _, d := range conf.Body {
				if d.Name == "node" {
					node := &Node{
						Code: d.Args[0], Name: d.Body.Get("name").Args[0],
						Address: d.Body.Get("address").Args[0],
						User:    d.Body.Get("user").Args[0],
						Password: "",
					}
					nodes = append(nodes, node)
				}
			}
		}
	}

	if config.Config.HasApi() {
		user, password := "", ""
		for authUserName, authPassword := range config.Config.Auth {
			user, password = authUserName, authPassword
			break
		}
		nodes = append(nodes, &Node{
			Code: "local", Name: "本地", User: user, Password: password,
			Address: "http://" + config.Config.Bind,
		})
	}

	return nodes
}

func (nc *nodeController) getNode(nodeName string) *Node {
	for _, node := range nc.list() {
		if node.Code == nodeName {
			return node
		}
	}
	return nil
}

func (nc *nodeController) add(ctx iris.Context) int {

	node := new(Node)
	err := ctx.ReadJSON(node)
	errors.PanicMessage(err, "解析节点信息错误")

	var conf *nginxConfig.Configuration
	if !files.Exists(nc.path) {
		conf = nginxConfig.New(nc.path)
	} else {
		conf, err = nginxConfig.Parse(nc.path, &nginxConfig.Options{})
		errors.PanicMessage(err, "解析配置文件")
	}

	nodes, err := query.Selects(conf, fmt.Sprintf("node('%s')", node.Code))

	if errors.IsNotFound(err) {
		nodeConf := conf.AddBody("node", node.Code)
		{
			nodeConf.AddBody("name", node.Name)
			nodeConf.AddBody("address", node.Address)
			nodeConf.AddBody("user", node.User)
			nodeConf.AddBody("password", node.Password)
		}
	} else {
		nodes[0].Body = []*nginxConfig.Directive{
			nginxConfig.New("name", node.Name),
			nginxConfig.New("address", node.Address),
			nginxConfig.New("user", node.User),
			nginxConfig.New("password", node.Password),
		}
	}

	err = ioutil.WriteFile(nc.path, conf.BodyBytes(), 0666)
	errors.PanicMessage(err, "写入文件配置错误")

	return iris.StatusNoContent
}

func (nc *nodeController) delete(ctx iris.Context) int {
	code := ctx.URLParam("code")
	logs.Info("删除节点：", code)

	conf, err := nginxConfig.Parse(nc.path, &nginxConfig.Options{})
	errors.PanicMessage(err, "解析配置文件")

	for i, d := range conf.Body {
		if d.Name == "node" && len(d.Args) == 1 && d.Args[0] == code {
			conf.Body = append(conf.Body[0:i], conf.Body[i+1:]...)
			break
		}
	}
	err = ioutil.WriteFile(nc.path, conf.BodyBytes(), 0666)
	errors.PanicMessage(err, "写入文件配置错误")
	return iris.StatusNoContent
}
