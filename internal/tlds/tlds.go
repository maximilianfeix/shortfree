// Package tlds is the single list of TLDs shortfree scans, plus the
// registration rules that decide which short labels can exist at all.
//
// Sources and verification dates for every entry are in docs/tlds.md.
package tlds

// Default is the list of TLDs scanned when none are given. The daily and the
// bootstrap workflow both build their job matrix from it ("shortfree tlds
// -json"), so adding a TLD here (plus its RDAP endpoint) is enough to get it
// scanned.
var Default = []string{
	"com", "net", "de", "ai", "io", "dev", "app", "tools",
	"co", "org", "xyz", "so", "to", "cc", "tv", "fm", "is", "ly", "run", "inc",
}

// Policy holds the registry rules relevant for 1-4 character names.
type Policy struct {
	// MinLen is the shortest label the registry accepts for new
	// registrations. Shorter names may exist (grandfathered), but they can
	// never become available, so they are neither scanned nor published.
	MinLen int
	// Canary is a label known to be registered. Before verifying a TLD the
	// scan looks it up; if the registry does not report it as registered,
	// its RDAP answers cannot be trusted that day (e.g. 404 for everything)
	// and the TLD's verification is skipped.
	Canary string
}

// policies lists TLDs that deviate from the default (every LDH label of
// length 1 and up is registrable, canary "nic").
var policies = map[string]Policy{
	"ly": {MinLen: 4}, // Libya NIC: 4-63 characters (registrar docs, see docs/tlds.md)
	"so": {MinLen: 3}, // .SO registry: 3-63 characters
	"to": {MinLen: 3}, // Tonic: 3-63 characters per major registrars
}

// Get returns the policy of tld with defaults filled in.
func Get(tld string) Policy {
	p := policies[tld]
	if p.MinLen < 1 {
		p.MinLen = 1
	}
	if p.Canary == "" {
		p.Canary = "nic"
	}
	return p
}

// Allows reports whether label can be registered under the policy.
func (p Policy) Allows(label string) bool { return len(label) >= p.MinLen }

// Allowed reports whether label.tld can be registered.
func Allowed(tld, label string) bool { return Get(tld).Allows(label) }
