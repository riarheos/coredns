package ipsetter

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("ipsetter", setup) }

func setup(c *caddy.Controller) error {
	setname, matchers, err := ipsetterParse(c)
	if err != nil {
		return plugin.Error("ipsetter", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return &IPSetter{Next: next, SetName: setname, Matchers: matchers}
	})

	return nil
}

func ipsetterParse(c *caddy.Controller) (string, []domainMatcher, error) {
	result := make([]domainMatcher, 0)
	setName := ""
	for c.Next() {
		ra := c.RemainingArgs()
		if len(ra) != 1 {
			return "", nil, c.ArgErr()
		}
		setName = ra[0]

		for c.NextBlock() {
			v := c.Val()
			if len(v) == 0 {
				return "", nil, c.ArgErr()
			}

			if v[0] == '*' {
				result = append(result, domainMatcher{
					exact:  false,
					domain: v[1:] + ".",
				})
			} else {
				result = append(result, domainMatcher{
					exact:  true,
					domain: v + ".",
				})
			}
		}
	}
	return setName, result, nil
}
