package config

import (
	"bytes"
	"fmt"
	cfg "github.com/ihaiker/aginx/v2/core/nginx/config"
	"github.com/ihaiker/aginx/v2/core/util"
	"net/url"
	"strconv"
	"strings"
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
			for _, body := range d.Body {
				switch body.Name {
				case "mode":
					c.Auth.Mode = body.Args[0]
				case "users":
					for _, auth := range body.Args {
						name, passwd := util.Split2(auth, "=")
						if passwd != "" {
							c.Auth.Users[name] = passwd
						}
					}
					for _, uap := range body.Body {
						c.Auth.Users[uap.Name] = strings.Join(uap.Args, "")
					}
				case "ldap":
					for _, ldapItem := range body.Body {
						switch ldapItem.Name {
						case "server":
							c.Auth.LDAP.Server = strings.Join(ldapItem.Args, "")
						case "bindDn", "bind-dn", "bind_dn":
							c.Auth.LDAP.BindDn = strings.Join(ldapItem.Args, "")
						case "password":
							c.Auth.LDAP.Password = strings.Join(ldapItem.Args, "")
						case "baseDn", "base-dn", "base_dn":
							c.Auth.LDAP.BaseDn = strings.Join(ldapItem.Args, "")
						case "usernameAttribute", "username-attribute",
							"username_attribute":
							c.Auth.LDAP.UsernameAttribute = strings.Join(ldapItem.Args, "")
						case "filter":
							c.Auth.LDAP.Filter = strings.Join(ldapItem.Args, "")
						case "tlsCa", "tls-ca", "tls_ca":
							c.Auth.LDAP.TLSCa = strings.Join(ldapItem.Args, "")
						case "tlsCert", "tls-cert", "tls_cert":
							c.Auth.LDAP.TLSCert = strings.Join(ldapItem.Args, "")
						case "tlsKey", "tls-key", "tls_key":
							c.Auth.LDAP.TLSKey = strings.Join(ldapItem.Args, "")
						}
					}
				}
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
		case "backup":
			for _, ld := range d.Body {
				switch ld.Name {
				case "dir":
					c.Backup.Dir = ld.Args[0]
				case "limit":
					c.Backup.Limit, _ = strconv.Atoi(ld.Args[0])
				case "day-limit", "dayLimit", "day_limit":
					c.Backup.DayLimit, _ = strconv.Atoi(ld.Args[0])
				case "crontab":
					c.Backup.Crontab = ld.Args[0]
				}
			}
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
