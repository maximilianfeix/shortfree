package main

import (
	"context"
	"fmt"
	"math/bits"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/maximilianfeix/shortfree/internal/names"
)

// version is set by the release build (-ldflags "-X main.version=...").
var version = ""

func appVersion() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return strings.TrimPrefix(bi.Main.Version, "v")
	}
	return "dev"
}

const mcpInstructions = `shortfree knows every unregistered 1-4 character domain on .com .net .de .ai .io .dev .app .tools (rescanned daily) and can check any domain live.

Workflow: use find_free_domains or suggest_domains to discover names from the daily snapshot, then confirm the ones the user cares about with check_domains (live DNS + registry RDAP) before calling them free. Statuses: rdap_verified = the registry's RDAP said "not registered" at the last scan; unverified = not in DNS, RDAP check still pending; dns_only = 4 character name not in DNS (never RDAP-checked, may be registered without name servers); dropping = registered but in redemption/pending delete. premium_likely = 1-2 character name outside .com/.net/.de, usually reserved or premium-priced by the registry even when RDAP says available. Final price and availability are always decided by the registrar.`

// mcpServer holds the shared state of all tools.
type mcpServer struct {
	data *dataSource
	chk  *checker
}

func cmdMCP(ctx context.Context, args []string) error {
	fs := newFlags("mcp")
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), `Usage: shortfree mcp

Runs an MCP server on stdin/stdout with the tools check_domains,
find_free_domains, suggest_domains and domain_stats.

Environment:
  SHORTFREE_DATA_URL  base URL of the published data (default `+defaultDataURL+`)
`)
	}
	fs.Parse(args)
	s := &mcpServer{data: newDataSource(), chk: newChecker()}
	return s.server().Run(ctx, &mcp.StdioTransport{})
}

func readOnly(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{Title: title, ReadOnlyHint: true, IdempotentHint: true}
}

func (s *mcpServer) server() *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:       "shortfree",
		Title:      "shortfree – short domain finder",
		Version:    appVersion(),
		WebsiteURL: "https://maximilianfeix.github.io/shortfree/",
	}, &mcp.ServerOptions{Instructions: mcpInstructions})

	mcp.AddTool(srv, &mcp.Tool{
		Name: "check_domains",
		Description: `Live availability check for up to 50 domains (any length, any TLD, e.g. "q7.tools", "acme.io"). ` +
			`For each domain it asks the TLD's own authoritative name servers (NS query, no resolver cache) and the registry's RDAP service, ` +
			`so the answer is current to the second, unlike the daily snapshot used by find_free_domains. ` +
			`Returns per domain: dns (free|taken|unknown), rdap (available|registered|dropping|error, "-" if the TLD has no RDAP endpoint in shortfree), ` +
			`verdict ("FREE" = registry confirms not registered; "taken"; "dropping soon"; "probably free (...)" when only DNS answered), ` +
			`free (true only for FREE) and premium_likely (1-2 character names outside .com/.net/.de, which registries usually reserve or sell at premium prices although RDAP reports them as unregistered). ` +
			`RDAP works for .com .net .de .ai .io .dev .app .tools; other TLDs get a DNS-only verdict. Use this before telling a user a domain is free. Takes about 1 s per domain (rate-limited per registry).`,
		Annotations: readOnly("Check domains live"),
	}, s.checkDomains)

	mcp.AddTool(srv, &mcp.Tool{
		Name: "find_free_domains",
		Description: `Search the daily snapshot of unregistered 1-4 character domains on .com .net .de .ai .io .dev .app .tools. ` +
			`Filters: tld (one or comma-separated, default all), length (1-4, default all, shortest first), kind (letters|digits|mixed|any), contains, starts_with, include_dropping. ` +
			`Each result has a status: rdap_verified (1-3 chars, registry RDAP said not registered at the last scan), unverified (1-3 chars, not in DNS, RDAP pending), ` +
			`dns_only (4 chars: not in the zone, never RDAP-checked; may be registered without name servers) or dropping (registered, in redemption/pending delete, with expiry; only with include_dropping). ` +
			`Results are ordered by length, then status, then TLD (com net de ai io dev app tools), then a-z before 0-9. total_matches counts all hits even when limit (default 100, max 1000) cuts the list. ` +
			`The snapshot is up to a day old: confirm picks with check_domains.`,
		Annotations: readOnly("Find free short domains"),
	}, s.findFreeDomains)

	mcp.AddTool(srv, &mcp.Tool{
		Name: "suggest_domains",
		Description: `Turn a word, brand or project idea into short (max 4 character) domain names that are free on .com .net .de .ai .io .dev .app .tools, best first. ` +
			`Deterministic heuristic, scored: the word itself if <= 4 chars (100), initials of a multi-word idea (95), a short word of the idea (90), initials of the first two words (85), first 4/3/2 letters (80/75/70), ` +
			`consonant skeleton without vowels like flickr -> flckr (78/73), two-word blends 2+2, 1+3, 3+1 letters (72/68), then startup patterns on those bases: ` +
			`add a digit 1-9 (55-47), add x (50), add hq or ly (48), go/my prefix (45). Stop words (the, of, and, ...) are dropped, accents transliterated, camelCase split. ` +
			`Each candidate is looked up in the daily snapshot per TLD; 1-2 character names on TLDs other than .com/.net/.de lose 15 points because they are usually premium. ` +
			`Ties go to the TLD order given in tlds (default com net de ai io dev app tools), then to RDAP-verified over unverified/dns_only. ` +
			`Longer names are out of scope: check those with check_domains. Confirm favourites with check_domains before registering.`,
		Annotations: readOnly("Suggest short domains"),
	}, s.suggestDomains)

	mcp.AddTool(srv, &mcp.Tool{
		Name: "domain_stats",
		Description: `Counts of free short domains per TLD and label length (1-4) from the latest daily scan: rdap_verified, unverified (DNS only, RDAP pending), ` +
			`dropping (in redemption/pending delete), dns_only_4char (4 character names not in the zone) and valid_names (how many LDH labels of that length exist). ` +
			`Use it to answer "how many 3 letter .io domains are left?" or to pick a TLD with the most free names before searching. Optional tld filter.`,
		Annotations: readOnly("Free domain statistics"),
	}, s.domainStats)
	return srv
}

// ---- check_domains ----

type checkInput struct {
	Domains []string `json:"domains" jsonschema:"1-50 domains like q7.tools or acme.io (label.tld, no subdomains)"`
}

type checkOutput struct {
	CheckedAt time.Time     `json:"checked_at"`
	Results   []checkResult `json:"results"`
}

func (s *mcpServer) checkDomains(ctx context.Context, _ *mcp.CallToolRequest, in checkInput) (*mcp.CallToolResult, checkOutput, error) {
	if len(in.Domains) == 0 || len(in.Domains) > 50 {
		return nil, checkOutput{}, fmt.Errorf("domains: give 1-50 domains, got %d", len(in.Domains))
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	return nil, checkOutput{CheckedAt: time.Now().UTC().Truncate(time.Second), Results: s.chk.checkAll(ctx, in.Domains)}, nil
}

// ---- find_free_domains ----

type findInput struct {
	TLD             string `json:"tld,omitempty" jsonschema:"TLD without or with dot, comma-separated for several (io,dev). Empty = all: com net de ai io dev app tools"`
	Length          int    `json:"length,omitempty" jsonschema:"exact label length 1-4, not counting the TLD. 0 = all lengths"`
	Kind            string `json:"kind,omitempty" jsonschema:"letters (a-z only), digits (0-9 only), mixed (letters+digits or hyphen) or any (default)"`
	Contains        string `json:"contains,omitempty" jsonschema:"only labels that contain this text, e.g. ai"`
	StartsWith      string `json:"starts_with,omitempty" jsonschema:"only labels that start with this text"`
	IncludeDropping bool   `json:"include_dropping,omitempty" jsonschema:"also return registered 1-3 character names in redemption or pending delete (status dropping)"`
	Limit           int    `json:"limit,omitempty" jsonschema:"maximum results, default 100, max 1000"`
}

type foundDomain struct {
	Domain        string     `json:"domain"`
	Status        string     `json:"status" jsonschema:"rdap_verified, unverified, dns_only or dropping"`
	PremiumLikely bool       `json:"premium_likely,omitempty"`
	Expires       *time.Time `json:"expires,omitempty"`
}

type findOutput struct {
	DataUpdated  time.Time     `json:"data_updated"`
	TotalMatches int           `json:"total_matches"`
	Returned     int           `json:"returned"`
	Domains      []foundDomain `json:"domains"`
	Notes        []string      `json:"notes,omitempty"`
}

func splitList(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' })
}

func (s *mcpServer) findFreeDomains(ctx context.Context, _ *mcp.CallToolRequest, in findInput) (*mcp.CallToolResult, findOutput, error) {
	var out findOutput
	limit := in.Limit
	switch {
	case limit == 0:
		limit = 100
	case limit < 0 || limit > 1000:
		return nil, out, fmt.Errorf("limit must be 1-1000")
	}
	if in.Length < 0 || in.Length > 4 {
		return nil, out, fmt.Errorf("length must be 1-4 (or 0 for all)")
	}
	kind := strings.ToLower(in.Kind)
	switch kind {
	case "", "any":
		kind = ""
	case "letters", "digits", "mixed":
	default:
		return nil, out, fmt.Errorf("kind must be letters, digits, mixed or any")
	}
	contains, starts := strings.ToLower(in.Contains), strings.ToLower(in.StartsWith)
	match := func(l string) bool {
		return (kind == "" || names.Kind(l) == kind) && strings.Contains(l, contains) && strings.HasPrefix(l, starts)
	}

	snap, err := s.data.available(ctx)
	if err != nil {
		return nil, out, fmt.Errorf("loading published data: %w", err)
	}
	tlds, err := snap.pickTLDs(splitList(in.TLD))
	if err != nil {
		return nil, out, err
	}
	out.DataUpdated = snap.Updated
	out.Domains = []foundDomain{}
	lengths := []int{1, 2, 3, 4}
	if in.Length > 0 {
		lengths = []int{in.Length}
	}
	premium := false
	emit := func(f foundDomain) {
		out.TotalMatches++
		if len(out.Domains) < limit {
			premium = premium || f.PremiumLikely
			out.Domains = append(out.Domains, f)
		}
	}

	for _, n := range lengths {
		if n == 4 {
			space := names.Space(4)
			for _, tld := range tlds {
				bm, err := s.data.bitmap(ctx, tld)
				if err != nil {
					out.Notes = append(out.Notes, fmt.Sprintf(".%s: 4 character data unavailable (%v)", tld, err))
					continue
				}
				if bm == nil {
					out.Notes = append(out.Notes, fmt.Sprintf(".%s: no 4 character data published yet", tld))
					continue
				}
				for bi, x := range bm {
					for ; x != 0; x &= x - 1 {
						i := bi*8 + bits.TrailingZeros8(x)
						if i >= space {
							break
						}
						if l := names.Name(i, 4); names.Valid(l) && match(l) {
							emit(foundDomain{Domain: l + "." + tld, Status: stDNSOnly})
						}
					}
				}
			}
			continue
		}
		type hit struct {
			foundDomain
			rank, tld, idx int
		}
		var hits []hit
		for ti, tld := range tlds {
			l := snap.TLDs[tld]
			add := func(label, status string, exp time.Time) {
				if len(label) != n || !match(label) {
					return
				}
				f := foundDomain{Domain: label + "." + tld, Status: status, PremiumLikely: premiumLikely(label, tld)}
				if !exp.IsZero() {
					f.Expires = &exp
				}
				hits = append(hits, hit{f, statusRank(status), ti, names.Index(label)})
			}
			for _, x := range l.Available {
				add(x, stVerified, time.Time{})
			}
			for _, x := range l.Unverified {
				add(x, stUnverified, time.Time{})
			}
			if in.IncludeDropping {
				for _, d := range l.Dropping {
					add(d.Name, stDropping, d.Expires)
				}
			}
		}
		sort.Slice(hits, func(i, j int) bool {
			a, b := hits[i], hits[j]
			if a.rank != b.rank {
				return a.rank < b.rank
			}
			if a.tld != b.tld {
				return a.tld < b.tld
			}
			return a.idx < b.idx
		})
		for _, h := range hits {
			emit(h.foundDomain)
		}
	}
	out.Returned = len(out.Domains)
	if out.TotalMatches > out.Returned {
		out.Notes = append(out.Notes, fmt.Sprintf("%d more matches: narrow the filters or raise limit (max 1000)", out.TotalMatches-out.Returned))
	}
	if premium {
		out.Notes = append(out.Notes, "premium_likely names are usually reserved or premium-priced by the registry; the registrar shows the real price")
	}
	out.Notes = append(out.Notes, fmt.Sprintf("snapshot from %s UTC; confirm with check_domains before registering", snap.Updated.UTC().Format("2006-01-02 15:04")))
	return nil, out, nil
}

// ---- suggest_domains ----

type suggestInput struct {
	Idea      string   `json:"idea" jsonschema:"a word, brand or project idea, e.g. Flicker or deep mind labs"`
	TLDs      []string `json:"tlds,omitempty" jsonschema:"TLDs to consider in order of preference, e.g. [io, dev]. Empty = com net de ai io dev app tools"`
	MaxLength int      `json:"max_length,omitempty" jsonschema:"longest label to suggest, 1-4, default 4"`
	Limit     int      `json:"limit,omitempty" jsonschema:"maximum suggestions, default 20, max 100"`
}

type suggestion struct {
	Domain        string `json:"domain"`
	Reason        string `json:"reason"`
	Score         int    `json:"score"`
	Status        string `json:"status" jsonschema:"rdap_verified, unverified or dns_only"`
	PremiumLikely bool   `json:"premium_likely,omitempty"`
}

type suggestOutput struct {
	Idea                 string       `json:"idea"`
	DataUpdated          time.Time    `json:"data_updated"`
	CandidatesConsidered int          `json:"candidates_considered"`
	Suggestions          []suggestion `json:"suggestions"`
	Notes                []string     `json:"notes,omitempty"`
}

func (s *mcpServer) suggestDomains(ctx context.Context, _ *mcp.CallToolRequest, in suggestInput) (*mcp.CallToolResult, suggestOutput, error) {
	out := suggestOutput{Idea: in.Idea, Suggestions: []suggestion{}}
	limit := in.Limit
	switch {
	case limit == 0:
		limit = 20
	case limit < 0 || limit > 100:
		return nil, out, fmt.Errorf("limit must be 1-100")
	}
	if in.MaxLength < 0 || in.MaxLength > 4 {
		return nil, out, fmt.Errorf("max_length must be 1-4")
	}
	if len(in.Idea) > 200 {
		return nil, out, fmt.Errorf("idea is too long (max 200 characters)")
	}
	cands := suggestCandidates(in.Idea, in.MaxLength)
	if len(cands) == 0 {
		return nil, out, fmt.Errorf("idea must contain letters or digits")
	}
	snap, err := s.data.available(ctx)
	if err != nil {
		return nil, out, fmt.Errorf("loading published data: %w", err)
	}
	tlds, err := snap.pickTLDs(in.TLDs)
	if err != nil {
		return nil, out, err
	}
	out.DataUpdated = snap.Updated
	out.CandidatesConsidered = len(cands)

	type ranked struct {
		suggestion
		rank, tld int
	}
	var all []ranked
	missing4 := map[string]bool{}
	for ti, tld := range tlds {
		for _, c := range cands {
			status := ""
			if len(c.Label) <= 3 {
				if st := snap.status[tld][c.Label]; st == stVerified || st == stUnverified {
					status = st
				}
			} else if !missing4[tld] {
				bm, err := s.data.bitmap(ctx, tld)
				if err != nil || bm == nil {
					missing4[tld] = true
					continue
				}
				if bm.Get(names.Index(c.Label)) {
					status = stDNSOnly
				}
			}
			if status == "" {
				continue
			}
			p := premiumLikely(c.Label, tld)
			r := c.Score
			if p {
				r -= 15
			}
			all = append(all, ranked{suggestion{c.Label + "." + tld, c.Reason, c.Score, status, p}, r, ti})
		}
	}
	sort.Slice(all, func(i, j int) bool {
		a, b := all[i], all[j]
		if a.rank != b.rank {
			return a.rank > b.rank
		}
		if a.tld != b.tld {
			return a.tld < b.tld
		}
		if statusRank(a.Status) != statusRank(b.Status) {
			return statusRank(a.Status) < statusRank(b.Status)
		}
		return a.Domain < b.Domain
	})
	for i := 0; i < len(all) && i < limit; i++ {
		out.Suggestions = append(out.Suggestions, all[i].suggestion)
	}
	for _, tld := range tlds {
		if missing4[tld] {
			out.Notes = append(out.Notes, fmt.Sprintf(".%s: 4 character data not available, only 1-3 character names checked", tld))
		}
	}
	if len(out.Suggestions) == 0 {
		out.Notes = append(out.Notes, "none of the short candidates is free on these TLDs; try other TLDs, a synonym, or check longer names with check_domains")
	}
	out.Notes = append(out.Notes, "snapshot data: confirm favourites with check_domains before registering")
	return nil, out, nil
}

// ---- domain_stats ----

type statsInput struct {
	TLD string `json:"tld,omitempty" jsonschema:"TLD or comma-separated TLDs; empty = all"`
}

type lengthCounts struct {
	RDAPVerified int `json:"rdap_verified"`
	Unverified   int `json:"unverified"`
	Dropping     int `json:"dropping"`
	DNSOnly4     int `json:"dns_only_4char"`
	ValidNames   int `json:"valid_names"`
}

type tldCounts struct {
	TLD     string                  `json:"tld"`
	Lengths map[string]lengthCounts `json:"lengths" jsonschema:"keyed by label length 1-4"`
	Free    int                     `json:"free_total" jsonschema:"rdap_verified + unverified + dns_only_4char"`
}

type statsOutput struct {
	DataUpdated time.Time   `json:"data_updated"`
	TLDs        []tldCounts `json:"tlds"`
}

func (s *mcpServer) domainStats(ctx context.Context, _ *mcp.CallToolRequest, in statsInput) (*mcp.CallToolResult, statsOutput, error) {
	var out statsOutput
	st, err := s.data.statsFile(ctx)
	if err != nil {
		return nil, out, fmt.Errorf("loading published stats: %w", err)
	}
	order := st.Order
	if len(order) == 0 {
		for t := range st.TLDs {
			order = append(order, t)
		}
		sort.Strings(order)
	}
	if req := splitList(in.TLD); len(req) > 0 {
		var picked []string
		for _, t := range req {
			t = strings.TrimPrefix(strings.ToLower(t), ".")
			if st.TLDs[t] == nil {
				return nil, out, fmt.Errorf("unsupported TLD %q: stats cover .%s", t, strings.Join(order, " ."))
			}
			picked = append(picked, t)
		}
		order = picked
	}
	out.DataUpdated = st.Updated
	out.TLDs = []tldCounts{}
	for _, t := range order {
		tc := tldCounts{TLD: t, Lengths: map[string]lengthCounts{}}
		for k, l := range st.TLDs[t] {
			if l == nil {
				continue
			}
			tc.Lengths[k] = lengthCounts{l.Available, l.Unverified, l.Dropping, l.Free, l.Valid}
			tc.Free += l.Available + l.Unverified + l.Free
		}
		out.TLDs = append(out.TLDs, tc)
	}
	return nil, out, nil
}
