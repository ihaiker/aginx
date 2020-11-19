package docker

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/ihaiker/aginx/v2/core/util"
	"net"
	"net/url"
)

func fromClient(config url.URL) (*client.Client, string, error) {
	opts := []client.Opt{client.WithAPIVersionNegotiation()}

	ip := ""
	if config.Path == "" { //tcp connect
		host := fmt.Sprintf("tcp://%s", config.Host)
		ip, _, _ = net.SplitHostPort(config.Host)
		opts = append(opts, client.WithHost(host))
	} else {
		host := fmt.Sprintf("unix:///%s%s", config.Host, config.Path)
		opts = append(opts, client.WithHost(host))
	}
	//用户指定了IP
	if paramIp := config.Query().Get("ip"); paramIp != "" {
		ip = paramIp
	}

	headers := map[string]string{}
	for _, header := range config.Query()["headers"] {
		name, value := util.Split2(header, "=")
		headers[name] = value
	}

	opts = append(opts, client.WithHTTPHeaders(headers))
	if _, has := config.Query()["tls"]; has {
		caPath := config.Query().Get("ca")
		certPath := config.Query().Get("cert")
		keyPath := config.Query().Get("key")
		opts = append(opts, client.WithTLSClientConfig(caPath, certPath, keyPath))
	}

	c, err := client.NewClientWithOpts(opts...)
	if ip == "" {
		return nil, "", fmt.Errorf("必须提供docker的ip参数")
	}
	return c, ip, err
}
