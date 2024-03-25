package notify

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
)

type HealthMonitors int8

const (
	HealthChecksIO HealthMonitors = iota
	BetterUptime
)

type HealthMonitor struct {
	ID        string
	Host      string
	Path      string
	Method    string
	withEndpt bool
}

var monitorParams = map[HealthMonitors]HealthMonitor{
	HealthChecksIO: {
		ID:        "NTFY_HEALTHCHECKS",
		Host:      "hc-ping.com",
		Path:      "",
		Method:    "POST",
		withEndpt: true,
	},
	BetterUptime: {
		ID:        "NTFY_BETTERUPTIME",
		Host:      "uptime.betterstack.com",
		Path:      "api/v1/heartbeat",
		Method:    "GET",
		withEndpt: false,
	},
}

func (hm HealthMonitors) getURL(endpoint string) *url.URL {
	return hm.getURLForID(os.Getenv(monitorParams[hm].ID), endpoint)
}

func (hm HealthMonitors) getURLForID(id string, endpoint string) *url.URL {
	host := monitorParams[hm].Host

	p := fmt.Sprintf("%s/%s", monitorParams[hm].Path, id)
	if monitorParams[hm].withEndpt {
		p = p + "/" + endpoint
	}
	p = path.Clean(p)

	return &url.URL{
		Scheme: "https",
		Host:   host,
		Path:   p,
	}
}

func (hm HealthMonitors) Ping(endpoint string, body *bytes.Buffer) (string, error) {
	url := hm.getURL(endpoint).String()

	resp, err := httpRequest(monitorParams[hm].Method, url, body, 10, 5, map[string]string{})
	if err != nil {
		return "", err
	}

	ret := resp.Status
	return ret, nil
}
