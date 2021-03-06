package ratelimit

import (
	"net"
	"strconv"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {

	caddy.RegisterPlugin("ratelimit", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {

	cfg := httpserver.GetConfig(c)

	rules, err := rateLimitParse(c)
	if err != nil {
		return err
	}

	rateLimit := RateLimit{Rules: rules}
	cfg.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		rateLimit.Next = next
		return rateLimit
	})

	return nil
}

func rateLimitParse(c *caddy.Controller) (rules []Rule, err error) {

	for c.Next() {
		var rule Rule

		args := c.RemainingArgs()
		switch len(args) {
		case 3:
			rule.Rate, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return rules, err
			}
			rule.Burst, err = strconv.Atoi(args[1])
			if err != nil {
				return rules, err
			}
			rule.Unit = args[2]
			for c.NextBlock() {
				if c.Val() == "allowLocalIPs" {
					for c.NextArg() {
						_, ipNet, err := net.ParseCIDR(c.Val())
						if err != nil {
							return rules, c.Errf("Cannot parse allow local ip")
						}
						rule.AllowIPs = append(rule.AllowIPs, ipNet)
					}
				}

				if c.Val() == "resources" {
					if !c.NextArg() {
						return rules, c.Errf("Missing method and resources")
					}
					method := c.Val()
					for c.NextArg() {
						rule.Resources = append(rule.Resources, Resource{Method: method, Url: c.Val()})
					}
				}
			}
		case 4:
			rule.Resources = append(rule.Resources, Resource{Method: "*", Url: args[0]})
			rule.Rate, err = strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return rules, err
			}
			rule.Burst, err = strconv.Atoi(args[2])
			if err != nil {
				return rules, err
			}
			rule.Unit = args[3]
		default:
			return rules, c.ArgErr()
		}

		rules = append(rules, rule)
	}

	return rules, nil
}
