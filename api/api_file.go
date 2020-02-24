package api

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

type aginxFile struct {
	*client
}

func (a aginxFile) Get(relativePath string) (string, error) {
	files, err := a.Search(relativePath)
	if err == nil {
		if content, has := files[relativePath]; has {
			return content, nil
		} else {
			return "", os.ErrNotExist
		}
	}
	return "", err
}

func (a aginxFile) New(relativePath, localFileAbsPath string) error {
	content, err := ioutil.ReadFile(localFileAbsPath)
	if err != nil {
		return err
	}
	return a.NewWithContent(relativePath, content)
}

func (a aginxFile) NewWithContent(relativePath string, content []byte) error {
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

	return a.request(http.MethodPost, "/file", buf, nil, func(r *http.Request) {
		r.Header.Set("Content-Type", writer.FormDataContentType())
	})
}

func (a aginxFile) Remove(relativePath string) error {
	return a.request(http.MethodDelete, "/file?file="+url.QueryEscape(relativePath), nil, nil)
}

func (a aginxFile) Search(relativePaths ...string) (files map[string]string, err error) {
	err = a.request(http.MethodGet, a.get("/file", relativePaths), nil, &files)
	return
}
