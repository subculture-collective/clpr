package twitch

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
)

// Fixture transport is usable only with literal disposable credentials and a
// loopback server. Release configuration separately rejects the setting.
func fixtureTransport(raw, clientID, secret string) (http.RoundTripper, error) {
	target, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(target.Hostname())
	if target.Scheme != "http" || ip == nil || !ip.IsLoopback() || target.User != nil || target.Path != "" || target.RawQuery != "" || target.Fragment != "" || clientID != "test_client_id" || secret != "test_client_secret" {
		return nil, fmt.Errorf("Twitch test fixture requires a loopback HTTP origin and disposable test credentials")
	}
	return fixtureRoundTripper{target: target}, nil
}

type fixtureRoundTripper struct{ target *url.URL }

func (t fixtureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != "api.twitch.tv" && req.URL.Host != "id.twitch.tv" {
		return nil, fmt.Errorf("unexpected fixture provider host")
	}
	clone := req.Clone(req.Context())
	u := *req.URL
	u.Scheme = t.target.Scheme
	u.Host = t.target.Host
	clone.URL = &u
	clone.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(clone)
}
