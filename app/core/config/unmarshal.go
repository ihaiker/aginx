package config

import (
	"bytes"
	"fmt"
	cfg "github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/util"
	"net/url"
)

func Unmarshal(data []byte, v interface{}) error {
	c := v.(*config)
	conf, err := cfg.ParseWith("aginx.conf", data, &cfg.Options{Delimiter: true, RemoveBrackets: true, RemoveAnnotation: true})

	if err != nil {
		return err
	}
	for _, d := range conf.Body {
		switch d.Name {
		case "log-level", "logLevel", "log_level":
			c.LogLevel = d.Args[0]
		case "log-file", "logFile", "log_file":
			c.LogFile = d.Args[0]
		case "bind":
			c.Bind = d.Args[0]
		case "api":
			c.Api = d.Args[0]
		case "allowIp", "allow_ip", "allow-ip":
			c.AllowIp = d.Args
		case "auth":
			c.Auth = make(map[string]string)
			for _, auth := range d.Args {
				name, passwd := util.Split2(auth, "=")
				if passwd == "" {
					return fmt.Errorf("error auth: %s", auth)
				}
				c.Auth[name] = passwd
			}
		case "expose":
			c.Expose = d.Args[0]
		case "disableAdmin", "disable-admin", "disable_admin":
			c.DisableAdmin = len(d.Args) == 0 || "true" == d.Args[0]
		case "disableDaemon", "disable_daemon", "disable-daemon":
			c.DisableDaemon = len(d.Args) == 0 || "true" == d.Args[0]
		case "disableApi", "disable-api", "disable_api":
			c.DisableApi = len(d.Args) == 0 || "true" == d.Args[0]

		case "nginx":
			c.Nginx = d.Args[0]

		case "registry":
			c.Registry = append(c.Registry, toConfigUrl(d))
		case "storage":
			c.Storage = toConfigUrl(d)
		case "cert", "certificate":
			c.Cert = append(c.Cert, toConfigUrl(d))
		case "node":
			//ignore
		default:
			return fmt.Errorf("未知配置: %s", d.Name)
		}
	}
	return nil
}

func toConfigUrl(d *cfg.Directive) string {
	out := bytes.NewBufferString(d.Args[0])
	for i, param := range d.Body {
		if i == 0 {
			out.WriteString("?")
		} else {
			out.WriteString("&")
		}
		for j, arg := range param.Args {
			if j != 0 {
				out.WriteString("&")
			}
			out.WriteString(url.QueryEscape(param.Name))
			out.WriteString("=")
			out.WriteString(arg)
		}
	}
	return out.String()
}
