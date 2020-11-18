package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

func (a *httpAginx) GetServers(filter *Filter) (server []*Server, err error) {
	c := a.http()
	uri := "/api/server"
	if filter != nil {
		uri += "?name=" + url.QueryEscape(filter.Name)
		uri += "&commit=" + url.QueryEscape(filter.Commit)
		uri += "&protocol=" + url.QueryEscape(string(filter.Protocol))
		if filter.ExactMatch {
			uri += "&exactMatch=true"
		}
	}
	server = make([]*Server, 0)
	err = c.request(http.MethodGet, uri, nil, &server)
	return
}

func (a *httpAginx) SetServer(server *Server) ([]string, error) {
	queries := make([]string, 0)
	c := a.http()
	out := bytes.NewBufferString("")
	if err := json.NewEncoder(out).Encode(server); err != nil {
		return nil, err
	}
	err := c.request(http.MethodPost, "/api/server", out, &queries)
	return queries, err
}

func (s *httpAginx) GetUpstream(filter *Filter) (upstreams []*Upstream, err error) {
	c := s.http()
	uri := "/api/upstream"
	if filter != nil {
		uri += "?name=" + url.QueryEscape(filter.Name)
		uri += "&commit=" + url.QueryEscape(filter.Commit)
		uri += "&protocol=" + url.QueryEscape(string(filter.Protocol))
		if filter.ExactMatch {
			uri += "&exactMatch=true"
		}
	}
	upstreams = make([]*Upstream, 0)
	err = c.request(http.MethodGet, uri, nil, &upstreams)
	return
}
func (s *httpAginx) SetUpstream(upstream *Upstream) ([]string, error) {
	queries := make([]string, 0)
	c := s.http()
	out := bytes.NewBufferString("")
	if err := json.NewEncoder(out).Encode(upstream); err != nil {
		return nil, err
	}
	err := c.request(http.MethodPost, "/api/upstream", out, &queries)
	return queries, err
}
