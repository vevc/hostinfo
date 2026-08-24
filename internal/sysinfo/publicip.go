package sysinfo

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// PublicIPInfo is the outbound public IP lookup result.
type PublicIPInfo struct {
	IP          string `json:"ip"`
	Source      string `json:"source"`
	LatencyMs   int64  `json:"latency_ms"`
	Reachable   bool   `json:"reachable"`
	Error       string `json:"error,omitempty"`
	AttemptedAt string `json:"attempted_at"`
}

var publicIPProviders = []struct {
	name string
	url  string
}{
	{name: "ipify", url: "https://api.ipify.org?format=text"},
	{name: "amazonaws", url: "https://checkip.amazonaws.com"},
	{name: "ifconfig.me", url: "https://ifconfig.me/ip"},
}

// CollectPublicIP probes external services to discover the server's public IP.
func CollectPublicIP() PublicIPInfo {
	started := time.Now().UTC()
	info := PublicIPInfo{
		AttemptedAt: started.Format(time.RFC3339),
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	var errors []string
	for _, provider := range publicIPProviders {
		reqStart := time.Now()
		req, err := http.NewRequest(http.MethodGet, provider.url, nil)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", provider.name, err))
			continue
		}
		req.Header.Set("User-Agent", "hostinfo/1.0")

		resp, err := client.Do(req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", provider.name, err))
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
		_ = resp.Body.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read body: %v", provider.name, err))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			errors = append(errors, fmt.Sprintf("%s: HTTP %d", provider.name, resp.StatusCode))
			continue
		}

		ip := strings.TrimSpace(string(body))
		if parsed := net.ParseIP(ip); parsed == nil {
			errors = append(errors, fmt.Sprintf("%s: invalid ip %q", provider.name, ip))
			continue
		}

		info.IP = ip
		info.Source = provider.name
		info.LatencyMs = time.Since(reqStart).Milliseconds()
		info.Reachable = true
		return info
	}

	info.Reachable = false
	if len(errors) > 0 {
		info.Error = strings.Join(errors, "; ")
	} else {
		info.Error = "no public IP providers available"
	}
	return info
}
