package util

import (
	"github.com/hashicorp/consul/api"
	"net/url"
	"strconv"
	"time"
)

func Consul(cfg url.URL) (*api.Config, error) {
	config := api.DefaultConfig()
	config.Address = cfg.Host
	config.Token = cfg.Query().Get("token")
	config.TokenFile = cfg.Query().Get("tokenFile")
	config.Datacenter = cfg.Query().Get("datacenter")
	config.Namespace = cfg.Query().Get("namespace")

	if waitTime := cfg.Query().Get("waitTime"); waitTime == "" {
		config.WaitTime = time.Second * 15
	} else if d, err := time.ParseDuration(waitTime); err != nil {
		return nil, err
	} else {
		config.WaitTime = d
	}

	if enableTLS, _ := strconv.ParseBool(cfg.Query().Get("tls")); enableTLS {
		config.Scheme = "https"
		config.TLSConfig = api.TLSConfig{
			CAFile:   cfg.Query().Get("ca"),
			CertFile: cfg.Query().Get("cert"),
			KeyFile:  cfg.Query().Get("key"),
		}
	}
	return config, nil
}
