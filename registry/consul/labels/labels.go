package consulLabels

import (
	"regexp"
	"strconv"
)

var keyRegexp = "aginx-domain"
var valueRegexp = regexp.MustCompile("([a-zA-Z0-9-_\\.]*)(,(weight=(\\d+)))?(,(ssl))?")

type Label struct {
	Domain  string
	Weight  int
	AutoSSL bool
}

func FindLabel(labs map[string]string) *Label {
	label, has := labs[keyRegexp]
	if !has {
		return nil
	}
	groups := valueRegexp.FindStringSubmatch(label)
	weight, _ := strconv.Atoi(groups[4])
	return &Label{
		Domain: groups[1], Weight: weight,
		AutoSSL: groups[6] == "ssl",
	}
}
