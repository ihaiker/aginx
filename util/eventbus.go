package util

import "github.com/asaskevich/EventBus"

var ebus = EventBus.New()

const (
	StorageFileChanged = "storage:file:changed"
	SSLExpire          = "ssl:expire"
)

func PublishFileChanged() {
	ebus.Publish(StorageFileChanged)
}

func PublishSSLExpire(domain string) {
	ebus.Publish(SSLExpire, domain)
}

func SubscribeFileChanged(fns ...func() error) {
	for _, fn := range fns {
		_ = ebus.Subscribe(StorageFileChanged, fn)
	}
}

func SubscribeSSLExpire(fns ...func(string)) {
	for _, fn := range fns {
		_ = ebus.Subscribe(SSLExpire, fn)
	}
}
