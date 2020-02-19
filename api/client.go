package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type client struct {
	address    string
	httpClient *http.Client
}

func (self *client) get(uri string, queries []string) string {
	if len(queries) > 0 {
		return uri + "?q=" + strings.Join(queries, "&q=")
	}
	return uri
}

func (self *client) response(resp *http.Response, ret interface{}) error {
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

func (self *client) request(method string, url string, body io.Reader, ret interface{}, extends ...func(r *http.Request)) error {
	if req, err := http.NewRequest(method, url, body); err != nil {
		return err
	} else {
		for _, extend := range extends {
			extend(req)
		}
		if resp, err := self.httpClient.Do(req); err != nil {
			return err
		} else {
			return self.response(resp, ret)
		}
	}
}
