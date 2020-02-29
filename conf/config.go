package conf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ihaiker/aginx/nginx"
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

func convert(cmd *cobra.Command, previousLayer string, directives []*nginx.Directive) (map[string]interface{}, error) {
	parameters := make(map[string]interface{})
	for _, directive := range directives {
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
	return parameters, nil
}
