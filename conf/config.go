package conf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ihaiker/aginx/nginx"
	"github.com/ihaiker/aginx/nginx/config"
	"github.com/ihaiker/aginx/plugins"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"strings"
)

func ReadConfig(configPath string, cmd *cobra.Command) error {
	if parameters, err := parse(cmd, configPath); err != nil {
		return err
	} else {
		viper.SetConfigType("json")
		data, _ := json.Marshal(parameters)
		return viper.ReadConfig(bytes.NewBuffer(data))
	}
}

func parse(cmd *cobra.Command, configPath string) (map[string]interface{}, error) {
	if content, err := ioutil.ReadFile(configPath); err != nil {
		return nil, err
	} else if cfg, err := nginx.ReaderReadable(nil, plugins.NewFile(configPath, content)); err != nil {
		return nil, err
	} else {
		return convert(cmd, "", cfg.Body)
	}
}

func key(pre, cur string) string {
	if pre == "" {
		return cur
	} else {
		return pre + "-" + cur
	}
}

func simpleServer(directive *config.Directive) ([]string, error) {
	services := make([]string, 0)
	for _, body := range directive.Body {
		domain := body.Name
		if len(body.Args) != 0 {
			return nil, fmt.Errorf("%s %s", domain, strings.Join(body.Args, " "))
		}
		ssl := false
		proxies := make([]string, 0)
		for _, d := range body.Body {
			if d.Name == "ssl" {
				ssl = true
			} else if d.Name == "server" {
				proxies = append(proxies, d.Args...)
			}
		}
		if len(proxies) == 0 {
			return nil, fmt.Errorf("No proxy address found：%s", domain)
		}
		if ssl {
			services = append(services, fmt.Sprintf("%s=ssl,%s", domain, strings.Join(proxies, ",")))
		} else {
			services = append(services, fmt.Sprintf("%s=%s", domain, strings.Join(proxies, ",")))
		}
	}
	return services, nil
}

func convert(cmd *cobra.Command, previousLayer string, directives []*config.Directive) (map[string]interface{}, error) {
	parameters := make(map[string]interface{})
	for _, directive := range directives {
		if directive.Name == "#" {
			//忽略注释
		} else if directive.Name == "server" {
			if servers, err := simpleServer(directive); err != nil {
				return nil, err
			} else {
				parameters["server"] = servers
			}
		} else {
			key := key(previousLayer, directive.Name)
			flag := cmd.PersistentFlags().Lookup(key)
			if flag == nil {
				return nil, fmt.Errorf("not flag found : %s.%s ", previousLayer, directive.Name)
			}

			if len(directive.Args) > 0 && len(directive.Body) > 0 {
				return nil, fmt.Errorf("error at : %s.%s ", previousLayer, directive.Name)
			} else if len(directive.Args) == 0 && len(directive.Body) == 0 {
				parameters[key] = "true"
			} else if len(directive.Args) > 0 {
				valueType := strings.ToLower(flag.Value.Type())
				if strings.HasSuffix(valueType, "array") || strings.HasSuffix(valueType, "slice") {
					parameters[key] = directive.Args
				} else if len(directive.Args) > 1 {
					return nil, fmt.Errorf(`error %s.%s Parameter up to one. (%s) `,
						previousLayer, directive.Name, strings.Join(directive.Args, " "))
				} else {
					parameters[key] = directive.Args[0]
				}
			} else {
				parameters[key] = "true"
				if subParams, err := convert(cmd, key, directive.Body); err != nil {
					return nil, err
				} else {
					for k, v := range subParams {
						parameters[k] = v
					}
				}
			}
		}
	}
	return parameters, nil
}
