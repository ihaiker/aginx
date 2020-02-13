package util

import "github.com/asaskevich/EventBus"

var EBus = EventBus.New()

const (
	StorageFileChanged = "storage:file:changed"
	SSLExpire          = "ssl:expire"
)
