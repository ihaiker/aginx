package api

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ihaiker/aginx/lego"
	"github.com/ihaiker/aginx/nginx/configuration"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func Queries(query ...string) []string {
	return query
}

type ApiError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
}

func (err *ApiError) Error() string {
	return err.Message
}

type Aginx interface {
	Auth(name, password string)

	//获取全局配置
	Configuration() *configuration.Configuration

	//nginx -s reload
	Reload() error

	NewFile(relativePath, localFileAbsPath string) error

	RemoveFile(relativePath string) error

	//查询 http.upstream，如果names参数存在将会命名在names参数中的upstream
	HttpUpstream(names ...string) ([]*configuration.Directive, error)

	//查询 http.server，如果names参数存在将会查找server_name在names中的server
	HttpServer(names ...string) ([]*configuration.Directive, error)

	//查询 stream.upstream，如果names参数存在将会命名在names参数中的upstream
	StreamUpstream(names ...string) ([]*configuration.Directive, error)

	//查询 stream.server，如果listens参数存在将会查找listen在listens中的server
	StreamServer(listens ...string) ([]*configuration.Directive, error)

	NewSsl(accountEmail, domain string) (*lego.StoreFile, error)

	ReNewSsl(domain string) (*lego.StoreFile, error)

	//查询配置
	Select(queries ...string) ([]*configuration.Directive, error)

	//添加配置
	Add(queries []string, addDirectives ...*configuration.Directive) error

	//删除
	Delete(queries ...string) error

	//更新配置
	Modify(queries []string, directive *configuration.Directive) error
}

type aginx struct {
	address string
	*http.Client
}

func New(address string, maker ...func(client *http.Client)) *aginx {
	apiUrl, _ := url.Parse(address)
	tp := &BaseAuthTransport{
		Transport: &http.Transport{},
	}
	if apiUrl.Scheme == "https" {
		tp.Transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{Transport: tp}
	client.Timeout = time.Second * 3
	for _, f := range maker {
		f(client)
	}
	return &aginx{
		Client: client, address: address,
	}
}

func (self *aginx) Auth(name, password string) {
	if tp, match := self.Client.Transport.(*BaseAuthTransport); match {
		tp.Name = name
		tp.Password = password
	}
}

func (self *aginx) Configuration() (*configuration.Configuration, error) {
	if directives, err := self.Select(); err != nil {
		return nil, err
	} else {
		return (*configuration.Configuration)(directives[0]), nil
	}
}

func (self *aginx) NewFile(relativePath, localFileAbsPath string) error {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	if formFile, err := writer.CreateFormFile("file", localFileAbsPath); err != nil {
		return err
	} else {
		if srcFile, err := os.Open(localFileAbsPath); err != nil {
			return err
		} else {
			defer func() { _ = srcFile.Close() }()
			if _, err = io.Copy(formFile, srcFile); err != nil {
				return err
			}
		}
	}
	if err := writer.WriteField("path", relativePath); err != nil {
		return err
	}
	_ = writer.Close()

	return self.request(http.MethodPost, "/file", buf, nil, func(r *http.Request) {
		r.Header.Set("Content-Type", writer.FormDataContentType())
	})
}

func (self *aginx) RemoveFile(relativePath string) error {
	file := base64.URLEncoding.EncodeToString([]byte(relativePath))
	return self.request(http.MethodDelete, "/file?file="+file, nil, nil)
}

func (self *aginx) HttpUpstream(names ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/http/upstream", names), nil, &directives)
	return
}

func (self *aginx) HttpServer(names ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/http/server", names), nil, &directives)
	return
}

func (self *aginx) StreamUpstream(names ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/stream/upstream", names), nil, &directives)
	return
}

func (self *aginx) StreamServer(listens ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/stream/server", listens), nil, &directives)
	return
}

func (self *aginx) NewSsl(accountEmail, domain string) (sf *lego.StoreFile, err error) {
	sf = new(lego.StoreFile)
	err = self.request(http.MethodPut, fmt.Sprintf("/ssl/%s?email=%s"+domain, accountEmail), nil, sf)
	return
}

func (self *aginx) ReNewSsl(domain string) (sf *lego.StoreFile, err error) {
	sf = new(lego.StoreFile)
	err = self.request(http.MethodPut, fmt.Sprintf("/ssl/%s"+domain), nil, sf)
	return
}

func (self *aginx) get(uri string, queries []string) string {
	if len(queries) > 0 {
		return uri + "?q=" + strings.Join(queries, "&q=")
	}
	return uri
}

func (self *aginx) response(resp *http.Response, ret interface{}) error {
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	if bs, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
			errApi := &ApiError{}
			if err := json.Unmarshal(bs, errApi); err == nil {
				return errApi
			}
		}
		return errors.New(string(bs))

	} else {
		return json.Unmarshal(bs, ret)
	}
}

func (self *aginx) request(method string, uri string, body io.Reader, ret interface{}, extends ...func(r *http.Request)) error {
	if req, err := http.NewRequest(method, self.address+uri, body); err != nil {
		return err
	} else {
		for _, extend := range extends {
			extend(req)
		}
		if resp, err := self.Client.Do(req); err != nil {
			return err
		} else {
			return self.response(resp, ret)
		}
	}
}

func (self *aginx) Select(queries ...string) (directives []*configuration.Directive, err error) {
	directives = make([]*configuration.Directive, 0)
	err = self.request(http.MethodGet, self.get("/api", queries), nil, &directives)
	return
}

func (self *aginx) Add(queries []string, addDirectives ...*configuration.Directive) error {
	if len(addDirectives) == 0 {
		return errors.New("addDirectives is empty")
	}
	body := bytes.NewBufferString("")
	for _, directive := range addDirectives {
		body.WriteString(directive.Pretty(0))
		body.WriteString("\n")
	}
	return self.request(http.MethodPut, self.get("/api", queries), body, nil)
}

func (self *aginx) Delete(queries ...string) error {
	return self.request(http.MethodDelete, self.get("/api", queries), nil, nil)
}

func (self *aginx) Modify(queries []string, directive *configuration.Directive) error {
	body := bytes.NewBufferString("")
	body.WriteString(directive.Pretty(0))
	return self.request(http.MethodPost, self.get("/api", queries), body, nil)
}

func (self *aginx) Reload() error {
	return self.request(http.MethodGet, "/reload", nil, nil)
}
