package api

import "net/http"

type BaseAuthTransport struct {
	Name     string
	Password string
	*http.Transport
}

func (bat *BaseAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if bat.Name != "" {
		req.SetBasicAuth(bat.Name, bat.Password)
	}
	return bat.Transport.RoundTrip(req)
}
