package docker

import (
	"github.com/ihaiker/aginx/logs"
	"regexp"
	"strconv"
)

var logger = logs.New("registor", "engine", "docker")

var keyRegexp = regexp.MustCompile("aginx.domain(\\.(\\d+))?")
var valueRegexp = regexp.MustCompile("([a-zA-Z0-9-_\\.]*)(,(weight=(\\d+)))?(,(internal))?")

func findLabels(labs map[string]string) labels {
	labels := labels{}
	for key, value := range labs {
		if keyRegexp.MatchString(key) && valueRegexp.MatchString(value) {

			domain := valueRegexp.FindStringSubmatch(value)
			port := keyRegexp.FindStringSubmatch(key)

			label := label{Domain: domain[1]}
			label.Weight, _ = strconv.Atoi(domain[4])
			label.Internal = domain[6] == "internal"
			label.Port, _ = strconv.Atoi(port[2])

			labels[label.Port] = label
		}
	}
	return labels
}
