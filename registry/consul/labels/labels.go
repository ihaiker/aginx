package consulLabels

import (
	"regexp"
	"strconv"
	"strings"
)

var keyRegexp = "aginx-domain"
var valueRegexp = regexp.MustCompile("([a-zA-Z0-9-_\\.]*)(,(weight=(\\d+)))?(,(ssl))?")

type Label struct {
	Domain  string
	Weight  int
	AutoSSL bool
}

func FindLabel(labs map[string]string) []*Label {
	label, has := labs[keyRegexp]
	if !has {
		return nil
	}
	values := strings.Split(label, ";")
	labels := make([]*Label, 0)
	for _, value := range values {
		groups := valueRegexp.FindStringSubmatch(value)
		weight, _ := strconv.Atoi(groups[4])
		labels = append(labels, &Label{
			Domain: groups[1], Weight: weight,
			AutoSSL: groups[6] == "ssl",
		})
	}
	return labels
}
