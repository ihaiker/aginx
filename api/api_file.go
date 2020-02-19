package api

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type aginxFile struct {
	*client
}

func (a aginxFile) New(relativePath, localFileAbsPath string) error {
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

	return a.request(http.MethodPost, "/file", buf, nil, func(r *http.Request) {
		r.Header.Set("Content-Type", writer.FormDataContentType())
	})
}

func (a aginxFile) Remove(relativePath string) error {
	file := base64.URLEncoding.EncodeToString([]byte(relativePath))
	return a.request(http.MethodDelete, "/file?file="+file, nil, nil)
}

func (a aginxFile) Search(queries ...string) (files [][]string, err error) {
	err = a.request(http.MethodGet, a.get("/file", queries), nil, &files)
	return
}
