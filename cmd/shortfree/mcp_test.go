package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/maximilianfeix/shortfree/internal/names"
	"github.com/maximilianfeix/shortfree/internal/probe"
	"github.com/maximilianfeix/shortfree/internal/rdap"
	"github.com/maximilianfeix/shortfree/internal/store"
)

var testUpdated = time.Date(2026, 10, 26, 3, 30, 0, 0, time.UTC)

// fakeData serves available.json, stats.json and a .dev 4 character bitmap
// (.io has none: 404) and counts requests per path.
func fakeData(t *testing.T) (*httptest.Server, map[string]*atomic.Int64) {
	t.Helper()
	avail := Available{Updated: testUpdated, TLDs: map[string]*TLDList{
		"com": {Available: []string{"q7x", "zz9"}, Dropping: []Dropping{}, Unverified: []string{}},
		"dev": {Available: []string{"q7", "dm", "dml", "abc", "x-1"}, Dropping: []Dropping{{"ab", testUpdated.AddDate(0, 0, 20)}}, Unverified: []string{"zq"}},
		"io":  {Available: []string{"dml", "q", "77"}, Dropping: []Dropping{}, Unverified: []string{"dmx"}},
	}}
	stats := Stats{Updated: testUpdated, Order: []string{"com", "dev", "io"}, TLDs: map[string]map[string]*LenStats{
		"com": {"1": {Valid: 36}, "2": {Valid: 1296}, "3": {Available: 2, Valid: 47952}, "4": {Free: 0, Valid: 1774224}},
		"dev": {"1": {Valid: 36}, "2": {Available: 2, Unverified: 1, Dropping: 1, Valid: 1296}, "3": {Available: 3, Valid: 47952}, "4": {Free: 4, Valid: 1774224}},
		"io":  {"1": {Available: 1, Valid: 36}, "2": {Available: 1, Valid: 1296}, "3": {Available: 1, Unverified: 1, Valid: 47952}, "4": {Valid: 1774224}},
	}}
	bm := store.NewBitmap(names.Space(4))
	for _, l := range []string{"abcd", "dmlx", "fl1r", "a-bc"} {
		bm.Set(names.Index(l), true)
	}
	files := map[string][]byte{"/data/4/dev.bin": bm}
	files["/data/available.json"], _ = json.Marshal(avail)
	files["/data/stats.json"], _ = json.Marshal(stats)
	hits := map[string]*atomic.Int64{}
	for p := range files {
		hits[p] = &atomic.Int64{}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		hits[r.URL.Path].Add(1)
		w.Write(b)
	}))
	t.Cleanup(srv.Close)
	return srv, hits
}

// fakeChecker answers from fixed tables instead of DNS and RDAP.
func fakeChecker() *checker {
	return &checker{
		dns: func(_ context.Context, label, tld string) (probe.Result, error) {
			switch label + "." + tld {
			case "free.dev", "q7.ai", "gone.io", "flaky.com", "nordap.gg":
				return probe.Free, nil
			case "broken.zz":
				return probe.Unknown, errors.New("lookup NS zz: no such host")
			}
			return probe.Taken, nil
		},
		rdap: func(_ context.Context, label, tld string) (rdap.Info, bool, error) {
			if _, ok := rdap.Endpoints[tld]; !ok {
				return rdap.Info{}, false, nil
			}
			switch label + "." + tld {
			case "free.dev", "q7.ai":
				return rdap.Info{Status: rdap.Available}, true, nil
			case "gone.io":
				return rdap.Info{Status: rdap.Dropping, Expires: testUpdated}, true, nil
			case "flaky.com":
				return rdap.Info{Status: rdap.Error}, true, errors.New("HTTP 503")
			}
			return rdap.Info{Status: rdap.Registered}, true, nil
		},
	}
}

func newTestServer(t *testing.T) (*mcpServer, map[string]*atomic.Int64) {
	srv, hits := fakeData(t)
	t.Setenv("SHORTFREE_DATA_URL", srv.URL+"/data")
	return &mcpServer{data: newDataSource(), chk: fakeChecker()}, hits
}

func domainsOf(f []foundDomain) []string {
	var out []string
	for _, d := range f {
		out = append(out, d.Domain+":"+d.Status)
	}
	return out
}

func TestCheckDomains(t *testing.T) {
	s, _ := newTestServer(t)
	_, out, err := s.checkDomains(context.Background(), nil, checkInput{Domains: []string{
		"Free.DEV.", "https://q7.ai/path", "google.com", "gone.io", "flaky.com", "nordap.gg", "broken.zz", "a.b.c", "-x.dev",
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		domain, dns, rdap, verdict string
		free, premium              bool
	}{
		{"free.dev", "free", "available", "FREE", true, false},
		{"q7.ai", "free", "available", "FREE", true, true},
		{"google.com", "taken", "registered", "taken", false, false},
		{"gone.io", "free", "dropping", "dropping soon", false, false},
		{"flaky.com", "free", "error", "probably free (RDAP failed)", false, false},
		{"nordap.gg", "free", "-", "probably free (no RDAP for this TLD)", false, false},
		{"broken.zz", "-", "-", "error", false, false},
		{"a.b.c", "-", "-", "invalid (want label.tld)", false, false},
		{"-x.dev", "-", "-", "invalid (want label.tld)", false, false},
	}
	if len(out.Results) != len(want) {
		t.Fatalf("got %d results", len(out.Results))
	}
	for i, w := range want {
		r := out.Results[i]
		if r.Domain != w.domain || r.DNS != w.dns || r.RDAP != w.rdap || r.Verdict != w.verdict || r.Free != w.free || r.PremiumLikely != w.premium {
			t.Errorf("result %d = %+v, want %+v", i, r, w)
		}
	}
	if out.Results[3].Expires == nil {
		t.Error("dropping name lost its expiry")
	}

	if _, _, err := s.checkDomains(context.Background(), nil, checkInput{}); err == nil {
		t.Error("empty domain list accepted")
	}
	if _, _, err := s.checkDomains(context.Background(), nil, checkInput{Domains: make([]string, 51)}); err == nil {
		t.Error("51 domains accepted")
	}
}

func TestFindFreeDomains(t *testing.T) {
	s, _ := newTestServer(t)
	ctx := context.Background()
	cases := []struct {
		name string
		in   findInput
		want string
	}{
		{"two letters on dev, verified first", findInput{TLD: ".dev", Length: 2},
			"dm.dev:rdap_verified q7.dev:rdap_verified zq.dev:unverified"},
		{"dropping on request", findInput{TLD: "dev", Length: 2, IncludeDropping: true},
			"dm.dev:rdap_verified q7.dev:rdap_verified zq.dev:unverified ab.dev:dropping"},
		{"all lengths, several TLDs, letters only", findInput{TLD: "io,dev", Kind: "letters", Contains: "d"},
			"dm.dev:rdap_verified dml.io:rdap_verified dml.dev:rdap_verified dmx.io:unverified abcd.dev:dns_only dmlx.dev:dns_only"},
		{"4 chars from the bitmap, missing .io bitmap noted", findInput{TLD: "dev,io", Length: 4},
			"abcd.dev:dns_only a-bc.dev:dns_only dmlx.dev:dns_only fl1r.dev:dns_only"},
		{"digits", findInput{Kind: "digits"}, "77.io:rdap_verified"},
		{"mixed + starts_with", findInput{Kind: "mixed", StartsWith: "q"},
			"q7.dev:rdap_verified q7x.com:rdap_verified"},
	}
	for _, c := range cases {
		_, out, err := s.findFreeDomains(ctx, nil, c.in)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := strings.Join(domainsOf(out.Domains), " "); got != c.want {
			t.Errorf("%s:\n got %s\nwant %s", c.name, got, c.want)
		}
		if !out.DataUpdated.Equal(testUpdated) {
			t.Errorf("%s: data_updated = %v", c.name, out.DataUpdated)
		}
	}

	_, out, _ := s.findFreeDomains(ctx, nil, findInput{TLD: "dev,io", Length: 4})
	if !strings.Contains(strings.Join(out.Notes, "\n"), ".io: no 4 character data") {
		t.Errorf("missing-bitmap note absent: %v", out.Notes)
	}

	_, out, _ = s.findFreeDomains(ctx, nil, findInput{Limit: 2})
	if out.Returned != 2 || out.TotalMatches <= 2 || !strings.Contains(strings.Join(out.Notes, " "), "more matches") {
		t.Errorf("limit: returned %d of %d, notes %v", out.Returned, out.TotalMatches, out.Notes)
	}
	if d := out.Domains[0]; d.Domain != "q.io" || !d.PremiumLikely {
		t.Errorf("first result %+v, want premium q.io", d)
	}

	for _, bad := range []findInput{{TLD: "xyz"}, {Length: 5}, {Kind: "emoji"}, {Limit: 1001}} {
		if _, _, err := s.findFreeDomains(ctx, nil, bad); err == nil {
			t.Errorf("%+v accepted", bad)
		}
	}
}

func TestDataCache(t *testing.T) {
	s, hits := newTestServer(t)
	ctx := context.Background()
	now := time.Now()
	s.data.now = func() time.Time { return now }
	for i := 0; i < 3; i++ {
		if _, _, err := s.findFreeDomains(ctx, nil, findInput{TLD: "dev", Length: 4}); err != nil {
			t.Fatal(err)
		}
	}
	if n := hits["/data/available.json"].Load(); n != 1 {
		t.Errorf("available.json fetched %d times within the TTL", n)
	}
	if n := hits["/data/4/dev.bin"].Load(); n != 1 {
		t.Errorf("dev.bin fetched %d times within the TTL", n)
	}
	now = now.Add(61 * time.Minute)
	s.findFreeDomains(ctx, nil, findInput{TLD: "dev", Length: 4})
	if n := hits["/data/available.json"].Load(); n != 2 {
		t.Errorf("available.json not refreshed after the TTL (%d fetches)", n)
	}

	// A failed refresh keeps serving the stale copy.
	s.data.base = "http://127.0.0.1:1/"
	now = now.Add(2 * time.Hour)
	if _, out, err := s.findFreeDomains(ctx, nil, findInput{TLD: "dev", Length: 2}); err != nil || len(out.Domains) == 0 {
		t.Errorf("stale data not served: %v", err)
	}
}

func TestSuggestDomains(t *testing.T) {
	s, _ := newTestServer(t)
	ctx := context.Background()
	_, out, err := s.suggestDomains(ctx, nil, suggestInput{Idea: "Deep Mind Labs", TLDs: []string{"io", "dev"}})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, sg := range out.Suggestions {
		got = append(got, sg.Domain)
	}
	// dml = initials (95) on both TLDs, io first as requested; dm = initials of the first two
	// words (85) is premium_likely on .dev (-15); dmx (unverified) and dmlx (4 chars, bitmap)
	// are "add x" (50), tie broken by TLD order.
	want := "dml.io dml.dev dm.dev dmx.io dmlx.dev"
	if strings.Join(got, " ") != want {
		t.Errorf("suggestions = %v, want %s", got, want)
	}
	if out.Suggestions[0].Reason != "initials" || out.Suggestions[0].Status != stVerified {
		t.Errorf("first suggestion %+v", out.Suggestions[0])
	}
	if out.CandidatesConsidered == 0 {
		t.Error("no candidates counted")
	}
	if _, _, err := s.suggestDomains(ctx, nil, suggestInput{Idea: "!!!"}); err == nil {
		t.Error("idea without letters accepted")
	}
	if _, _, err := s.suggestDomains(ctx, nil, suggestInput{Idea: "x", TLDs: []string{"xyz"}}); err == nil {
		t.Error("unknown TLD accepted")
	}
}

func TestDomainStats(t *testing.T) {
	s, _ := newTestServer(t)
	_, out, err := s.domainStats(context.Background(), nil, statsInput{TLD: "dev"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.TLDs) != 1 || out.TLDs[0].TLD != "dev" {
		t.Fatalf("tlds = %+v", out.TLDs)
	}
	d := out.TLDs[0]
	if d.Free != 2+1+3+4 || d.Lengths["2"].Dropping != 1 || d.Lengths["4"].DNSOnly4 != 4 || d.Lengths["3"].ValidNames != 47952 {
		t.Errorf("dev stats = %+v", d)
	}
	_, all, _ := s.domainStats(context.Background(), nil, statsInput{})
	if len(all.TLDs) != 3 || all.TLDs[0].TLD != "com" {
		t.Errorf("order = %+v", all.TLDs)
	}
	if _, _, err := s.domainStats(context.Background(), nil, statsInput{TLD: "xyz"}); err == nil {
		t.Error("unknown TLD accepted")
	}
}

// TestMCPRoundTrip talks to the server through the SDK's client over an
// in-memory transport, the way an assistant would.
func TestMCPRoundTrip(t *testing.T) {
	s, _ := newTestServer(t)
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.server().Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	if inst := cs.InitializeResult().Instructions; !strings.Contains(inst, "check_domains") {
		t.Errorf("instructions = %q", inst)
	}
	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var toolNames []string
	for _, tool := range tools.Tools {
		toolNames = append(toolNames, tool.Name)
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint || len(tool.Description) < 200 {
			t.Errorf("tool %s: weak metadata", tool.Name)
		}
	}
	if got := strings.Join(toolNames, ","); got != "check_domains,domain_stats,find_free_domains,suggest_domains" {
		t.Errorf("tools = %s", got)
	}

	call := func(name string, args map[string]any) *mcp.CallToolResult {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return res
	}
	res := call("find_free_domains", map[string]any{"tld": "dev", "length": 2})
	if res.IsError {
		t.Fatalf("find_free_domains failed: %+v", res.Content)
	}
	b, _ := json.Marshal(res.StructuredContent)
	var found findOutput
	if err := json.Unmarshal(b, &found); err != nil || found.Returned != 3 || found.Domains[0].Domain != "dm.dev" {
		t.Errorf("structured find result = %s (%v)", b, err)
	}
	if text := res.Content[0].(*mcp.TextContent).Text; !strings.Contains(text, "dm.dev") {
		t.Errorf("text content = %s", text)
	}

	res = call("check_domains", map[string]any{"domains": []string{"free.dev"}})
	b, _ = json.Marshal(res.StructuredContent)
	if !strings.Contains(string(b), `"verdict":"FREE"`) {
		t.Errorf("check_domains = %s", b)
	}

	res = call("find_free_domains", map[string]any{"kind": "emoji"})
	if !res.IsError {
		t.Error("invalid kind did not produce a tool error")
	}
	res = call("domain_stats", map[string]any{})
	if res.IsError {
		t.Errorf("domain_stats failed: %+v", res.Content)
	}
	res = call("suggest_domains", map[string]any{"idea": "deep mind labs"})
	if res.IsError {
		t.Errorf("suggest_domains failed: %+v", res.Content)
	}
}
