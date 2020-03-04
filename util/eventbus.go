package util

import "github.com/asaskevich/EventBus"

var ebus = EventBus.New()

const (
	StorageFileChanged = "storage:file:changed"
)

func PublishFileChanged() {
	ebus.Publish(StorageFileChanged)
}

func SubscribeFileChanged(fns ...func() error) {
	for _, fn := range fns {
		_ = ebus.Subscribe(StorageFileChanged, fn)
	}
}
