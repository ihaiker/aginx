package util

import "github.com/asaskevich/EventBus"

var EBus = EventBus.New()

const (
	NginxReload = "nginx:reload"
	SSLExpire   = "ssl:expire"
)

func init() {

}
