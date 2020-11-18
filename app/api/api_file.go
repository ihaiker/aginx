package api

import (
	"bytes"
	"github.com/ihaiker/aginx/v2/plugins/storage"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

type httpAginxFile struct {
	*client
}

func (a *httpAginxFile) New(relativePath, localFileAbsPath string) error {
	content, err := ioutil.ReadFile(localFileAbsPath)
	if err != nil {
		return err
	}
	return a.NewWithContent(relativePath, content)
}

func (a *httpAginxFile) Get(relativePath string) (file *storage.File, err error) {
	file = new(storage.File)
	err = a.client.request(http.MethodGet, "/api/file?q="+url.QueryEscape(relativePath), nil, file)
	return
}

func (a *httpAginxFile) NewWithContent(relativePath string, content []byte) error {
	buf := new(bytes.Buffer)

	writer := multipart.NewWriter(buf)
	if formFile, err := writer.CreateFormFile("file", relativePath); err != nil {
		return err
	} else if _, err = formFile.Write(content); err != nil {
		return err
	}
	if err := writer.WriteField("path", relativePath); err != nil {
		return err
	}
	_ = writer.Close()

	return a.request(http.MethodPost, "/api/file", buf, nil, func(r *http.Request) {
		r.Header.Set("Content-Type", writer.FormDataContentType())
	})
}

func (a *httpAginxFile) Remove(relativePath string) error {
	return a.request(http.MethodDelete, "/api/file?q="+url.QueryEscape(relativePath), nil, nil)
}

func (a *httpAginxFile) Search(relativePaths ...string) (files []*storage.File, err error) {
	err = a.request(http.MethodGet, a.get("/api/file/search", relativePaths), nil, &files)
	return
}
