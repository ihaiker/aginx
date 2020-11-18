package functions

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

func UpstreamName(domain string) string {
	name := strings.ReplaceAll(domain, ".", "_")
	name = strings.ReplaceAll(name, "*", "_x_")
	return snake(name)
}

// snake string, XxYy to xx_yy , XxYY to xx_yy
func snake(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

func camel(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func stringsFuncMap() template.FuncMap {
	return template.FuncMap{
		"upstreamName": UpstreamName,
		"join":         strings.Join,
		"format":       fmt.Sprintf,

		"replace":    strings.Replace,
		"replaceAll": strings.ReplaceAll,
		"hasSuffix":  strings.HasSuffix,
		"hasPrefix":  strings.HasPrefix,

		"index":     strings.Index,
		"lastIndex": strings.LastIndex,
		"indexStart": func(str, search string, idx int) (int, error) {
			if len(str) < idx-1 {
				return -1, errors.New("length not enough")
			}
			i := strings.Index(str[idx:], search)
			return i, nil
		},

		"toLower":  strings.ToLower,
		"toTitle":  strings.Title,
		"contains": strings.Contains,

		"trim":       strings.Trim,
		"trimSpace":  strings.TrimSpace,
		"trimLeft":   strings.TrimLeft,
		"trimRight":  strings.TrimRight,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,

		"parseBool":  strconv.ParseBool,
		"parseFloat": strconv.ParseFloat,
		"parseInt":   strconv.ParseInt,
		"parseUint":  strconv.ParseUint,

		"split":       strings.Split,
		"splitN":      strings.SplitN,
		"splitAfter":  strings.SplitAfter,
		"splitAfterN": strings.SplitAfterN,

		"snake": snake,

		// camel string, xx_yy to XxYy
		"camel": camel,

		"base64Decode": func(s string) (string, error) {
			v, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return "", err
			}
			return string(v), nil
		},
		"base64Encode": func(s string) (string, error) {
			return base64.StdEncoding.EncodeToString([]byte(s)), nil
		},
		"base64URLDecode": func(s string) (string, error) {
			v, err := base64.URLEncoding.DecodeString(s)
			if err != nil {
				return "", err
			}
			return string(v), nil
		},
		"base64URLEncode": func(s string) (string, error) {
			return base64.URLEncoding.EncodeToString([]byte(s)), nil
		},
		"parseJSON": func(s string) (interface{}, error) {
			if s == "" {
				return map[string]interface{}{}, nil
			}

			var data interface{}
			if err := json.Unmarshal([]byte(s), &data); err != nil {
				return nil, err
			}
			return data, nil
		},

		"regexReplaceAll": func(re, pl, s string) (string, error) {
			compiled, err := regexp.Compile(re)
			if err != nil {
				return "", err
			}
			return compiled.ReplaceAllString(s, pl), nil
		},

		"regexMatch": func(re, s string) (bool, error) {
			compiled, err := regexp.Compile(re)
			if err != nil {
				return false, err
			}
			return compiled.MatchString(s), nil
		},

		"toJSON": func(i interface{}) (string, error) {
			result, err := json.Marshal(i)
			if err != nil {
				return "", err
			}
			return string(bytes.TrimSpace(result)), err
		},
		"toJSONPretty": func(m map[string]interface{}) (string, error) {
			result, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				return "", err
			}
			return string(bytes.TrimSpace(result)), err
		},
	}
}
