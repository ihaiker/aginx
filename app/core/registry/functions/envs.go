package functions

import (
	"os"
	"strings"
	"text/template"
)

func envs() map[string]string {
	envs := map[string]string{}
	for _, env := range os.Environ() {
		kv := strings.SplitN(env, "=", 2)
		envs[kv[0]] = kv[1]
	}
	return envs
}

func hasEnv(env string) bool {
	val := os.Getenv(env)
	return val != ""
}

func env(key string) string {
	return os.Getenv(key)
}

func envTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"envs":   envs,
		"hasEnv": hasEnv,
		"env":    env,
	}
}
