package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type client struct {
	address    string
	httpClient *http.Client
}

func (self *client) get(uri string, queries []string) string {
	if len(queries) > 0 {
		values := url.Values{}
		for _, query := range queries {
			values.Add("q", query)
		}
		return uri + "?" + values.Encode()
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
				if errApi.Message == "file does not exist" {
					return os.ErrNotExist
				}
				return errApi
			}
		}
		return errors.New(string(bs))
	} else {
		return json.Unmarshal(bs, ret)
	}
}

func (self *client) request(method string, url string, body io.Reader, ret interface{}, extends ...func(r *http.Request)) error {
	if req, err := http.NewRequest(method, self.address+url, body); err != nil {
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
