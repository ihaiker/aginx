package consul

import (
	"fmt"
	"net/url"
	"strings"
)

type label struct {
	Domain   string
	AutoSSL  bool
	Template string
	Provider string
}

//aginx.domain.0=http://ws.renzhen.la
//aginx.domain.1=https://wss.renzhen.la
func findLabel(tags map[string]string) ([]*label, error) {
	if tags == nil || len(tags) == 0 {
		return nil, nil
	}
	labels := make([]*label, 0)
	for key, value := range tags {
		if !strings.HasPrefix(key, "aginx.domain") {
			continue
		}
		tag, err := url.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("error tag: %s", value)
		}

		label := new(label)
		label.Domain = tag.Host
		label.AutoSSL = tag.Scheme == "https"
		label.Template = tag.Query().Get("template")
		label.Provider = tag.Query().Get("provider")
		labels = append(labels, label)
	}
	return labels, nil
}
