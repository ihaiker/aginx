package api

import (
	"net/http"
	"net/url"
)

type httpAginxBackup struct {
	*client
}

func (h *httpAginxBackup) List() ([]*BackupFile, error) {
	backups := make([]*BackupFile, 0)
	err := h.request(http.MethodGet, "/api/backup", nil, &backups)
	return backups, err
}

func (h *httpAginxBackup) Delete(dateString string) error {
	return h.request(http.MethodDelete, "/api/backup?name="+dateString, nil, nil)
}

func (h *httpAginxBackup) Backup(comment string) (*BackupFile, error) {
	name := new(BackupFile)
	err := h.request(http.MethodPost, "/api/backup?comment="+url.QueryEscape(comment), nil, name)
	return name, err
}

func (h *httpAginxBackup) Rollback(dateString string) error {
	return h.request(http.MethodPut, "/api/backup?name="+dateString, nil, nil)
}
