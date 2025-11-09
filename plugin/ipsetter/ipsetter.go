package ipsetter

import (
	"context"
	"log"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/lrh3321/ipset-go"
	"github.com/miekg/dns"
)

type IPSetter struct {
	Next     plugin.Handler
	SetName  string
	Matchers []domainMatcher
}

type domainMatcher struct {
	exact  bool
	domain string
}

type resultingResponseWriter struct {
	dns.ResponseWriter
	msg *dns.Msg
}

func (w *resultingResponseWriter) WriteMsg(res *dns.Msg) error {
	w.msg = res
	return w.ResponseWriter.WriteMsg(res)
}

func (l *IPSetter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	m := false
	for _, q := range r.Question {
		if l.matches(q.Name) {
			m = true
			break
		}
	}
	if !m {
		return plugin.NextOrFailure(l.Name(), l.Next, ctx, w, r)
	}

	rw := &resultingResponseWriter{ResponseWriter: w, msg: r}
	res, err := plugin.NextOrFailure(l.Name(), l.Next, ctx, rw, r)

	if err == nil {
		for _, rr := range rw.msg.Answer {
			if rr.Header().Rrtype == dns.TypeA {
				rec := rr.(*dns.A)
				err = ipset.Add(l.SetName, &ipset.Entry{IP: rec.A, Replace: true})
				if err != nil {
					log.Printf("Error adding IP to set: %s", err)
				}
			}
		}
	}

	return res, err
}

func (l *IPSetter) matches(domain string) bool {
	for _, matcher := range l.Matchers {
		if matcher.exact {
			if matcher.domain == domain {
				return true
			}
		} else {
			if strings.HasSuffix(domain, matcher.domain) {
				return true
			}
		}
	}
	return false
}

func (l *IPSetter) Name() string { return "ipsetter" }
