// Package probe asks a TLD's authoritative name servers whether a name is
// delegated. NXDOMAIN from the registry's own servers means the name is not in
// the zone, which is the fast first filter; RDAP confirms afterwards.
package probe

import (
	"context"
	"fmt"
	"hash/fnv"
	"net"
	"sort"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/time/rate"
)

// Result is the zone status of one name.
type Result uint8

const (
	Unknown Result = iota // no usable answer after all retries
	Free                  // NXDOMAIN: not in the zone
	Taken                 // NOERROR: delegated (or otherwise present)
)

func (r Result) String() string {
	return [...]string{"unknown", "free", "taken"}[r]
}

// Prober queries one TLD's authoritative servers with a global rate limit.
type Prober struct {
	TLD     string
	servers []string
	limit   *rate.Limiter
	udp     *dns.Client
	tcp     *dns.Client
	retries int
}

// New resolves the TLD's name servers (IPv4 only, GitHub runners have no IPv6)
// and returns a prober limited to qps queries per second.
func New(ctx context.Context, tld string, qps int) (*Prober, error) {
	var r net.Resolver
	nss, err := r.LookupNS(ctx, tld+".")
	if err != nil {
		return nil, fmt.Errorf("lookup NS %s: %w", tld, err)
	}
	var servers []string
	for _, ns := range nss {
		ips, err := r.LookupIP(ctx, "ip4", ns.Host)
		if err != nil {
			continue
		}
		for _, ip := range ips {
			servers = append(servers, net.JoinHostPort(ip.String(), "53"))
		}
	}
	if len(servers) == 0 {
		return nil, fmt.Errorf("no reachable name servers for .%s", tld)
	}
	sort.Strings(servers)
	return &Prober{
		TLD:     tld,
		servers: servers,
		limit:   rate.NewLimiter(rate.Limit(qps), qps/10+1),
		udp:     &dns.Client{Net: "udp", Timeout: 2 * time.Second},
		tcp:     &dns.Client{Net: "tcp", Timeout: 4 * time.Second},
		retries: 5,
	}, nil
}

// Servers returns the server addresses in use.
func (p *Prober) Servers() []string { return p.servers }

// Check returns the zone status of label.TLD.
func (p *Prober) Check(ctx context.Context, label string) Result {
	fqdn := dns.Fqdn(label + "." + p.TLD)
	h := fnv.New32a()
	h.Write([]byte(fqdn))
	start := int(h.Sum32() % uint32(len(p.servers)))

	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeNS)
	m.RecursionDesired = false

	for attempt := 0; attempt < p.retries; attempt++ {
		if err := p.limit.Wait(ctx); err != nil {
			return Unknown
		}
		server := p.servers[(start+attempt)%len(p.servers)]
		r, _, err := p.udp.ExchangeContext(ctx, m, server)
		if err == nil && r.Truncated {
			r, _, err = p.tcp.ExchangeContext(ctx, m, server)
		}
		if err != nil || r == nil {
			backoff(ctx, attempt)
			continue
		}
		switch r.Rcode {
		case dns.RcodeNameError:
			return Free
		case dns.RcodeSuccess:
			return Taken
		}
		backoff(ctx, attempt)
	}
	return Unknown
}

func backoff(ctx context.Context, attempt int) {
	t := time.NewTimer(time.Duration(100<<attempt) * time.Millisecond)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
