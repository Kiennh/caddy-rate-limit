package ratelimit

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// RateLimit is an http.Handler that can limit request rate to specific paths or files
type RateLimit struct {
	Next  httpserver.Handler
	Rules []Rule
}

// Rule is a configuration for ratelimit
type Rule struct {
	Rate      int64
	Burst     int
	Resources []Resource
	Unit      string
	AllowIPs  []*net.IPNet
}

type Resource struct {
	Method string
	Url    string
}

const (
	ignorePrivateIP = true
	symbol          = "^"
)

var (
	caddyLimiter *CaddyLimiter
	localIpNets  []*net.IPNet
)

func init() {

	caddyLimiter = NewCaddyLimiter()
	// https://en.wikipedia.org/wiki/Private_network
	localCIDRs := []string{
		"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "::1/128", "fc00::/7",
	}
	for _, s := range localCIDRs {
		_, ipNet, err := net.ParseCIDR(s)
		if err == nil {
			localIpNets = append(localIpNets, ipNet)
		}
	}
}

func (rl RateLimit) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	retryAfter := time.Duration(0)

	// handle exception first
	for _, rule := range rl.Rules {
		for _, res := range rule.Resources {
			if strings.HasPrefix(res.Url, symbol) {
				res = Resource{Method: res.Method, Url: strings.TrimPrefix(res.Url, symbol)}
				if httpserver.Path(r.URL.Path).Matches(res.Url) && (res.Method == "*" || res.Method == r.Method) {
					return rl.Next.ServeHTTP(w, r)
				}
			}
		}
	}

	for _, rule := range rl.Rules {
		for _, res := range rule.Resources {
			if !httpserver.Path(r.URL.Path).Matches(res.Url) {
				continue
			}

			if r.Method != res.Method {
				continue
			}

			if ignorePrivateIP {
				// filter local ip address
				address, err := GetRemoteIP(r)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				if !IsLocalIpAddress(address, rule.AllowIPs) && IsLocalIpAddress(address, localIpNets) {
					continue
				}
			}

			sliceKeys := buildKeys(res.Url, r)
			for _, keys := range sliceKeys {
				ret := caddyLimiter.Allow(keys, rule)
				retryAfter = caddyLimiter.RetryAfter(keys)
				if !ret {
					w.Header().Add("X-RateLimit-RetryAfter", retryAfter.String())
					return http.StatusTooManyRequests, nil
				}
			}
		}
	}

	return rl.Next.ServeHTTP(w, r)
}
