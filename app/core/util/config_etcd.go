package util

import (
	"fmt"
	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/ihaiker/aginx/v2/core/util/errors"
	"net/url"
	"strconv"
	"time"
)

func Etcd(cfg url.URL) (*v3.Config, error) {
	config := &v3.Config{
		Endpoints: []string{cfg.Host},
	}

	if err := etcdConfigFromUrl(config, cfg); err != nil {
		return nil, err
	}
	return config, nil
}

func filedDuration(filed *time.Duration, name string, cfg url.URL) error {
	if value := cfg.Query().Get(name); value != "" {
		if d, err := time.ParseDuration(value); err != nil {
			return errors.Wrap(err, fmt.Sprintf("etcd config: %s=%s", name, value))
		} else {
			*filed = d
		}
	}
	return nil
}

func etcdConfigFromUrl(config *v3.Config, cfg url.URL) (err error) {
	defer errors.Catch(func(e error) {
		err = e
	})
	endpoints := cfg.Query()["endpoints"]
	if endpoints != nil && len(endpoints) != 0 {
		config.Endpoints = endpoints
	} else {
		config.Endpoints = []string{cfg.Host}
	}
	config.Username = cfg.Query().Get("username")
	config.Password = cfg.Query().Get("password")
	errors.PanicIfError(filedDuration(&config.AutoSyncInterval, "autoSyncInterval", cfg))
	errors.PanicIfError(filedDuration(&config.DialTimeout, "dialTimeout", cfg))
	errors.PanicIfError(filedDuration(&config.DialKeepAliveTime, "dialKeepAliveTime", cfg))
	errors.PanicIfError(filedDuration(&config.DialKeepAliveTimeout, "dialKeepAliveTimeout", cfg))

	if enableTLS, _ := strconv.ParseBool(cfg.Query().Get("tls")); enableTLS {
		tlsInfo := transport.TLSInfo{
			CertFile:      cfg.Query().Get("cert"),
			KeyFile:       cfg.Query().Get("key"),
			TrustedCAFile: cfg.Query().Get("ca"),
		}
		config.TLS, err = tlsInfo.ClientConfig()
	}
	return
}
