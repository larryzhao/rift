package pac

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return f
}

func TestParseDomains(t *testing.T) {
	f := writeFile(t, "domains.txt", `# comment
claude.ai

  anthropic.com
Claude.ai
!fast.example.com
! spaced.example.com
# trailing
`)
	proxy, direct, err := parseDomains(f)
	if err != nil {
		t.Fatalf("parseDomains: %v", err)
	}

	wantProxy := map[string]bool{"claude.ai": true, "anthropic.com": true}
	if len(proxy) != len(wantProxy) {
		t.Fatalf("proxy = %v, want %d entries", proxy, len(wantProxy))
	}
	for _, d := range proxy {
		if !wantProxy[d] {
			t.Errorf("unexpected proxy domain %q", d)
		}
	}

	wantDirect := map[string]bool{"fast.example.com": true, "spaced.example.com": true}
	for _, d := range direct {
		if !wantDirect[d] {
			t.Errorf("unexpected direct domain %q", d)
		}
	}
	if len(direct) != len(wantDirect) {
		t.Errorf("direct = %v, want %d entries", direct, len(wantDirect))
	}
}

func TestParseGFWList(t *testing.T) {
	plain := "[AutoProxy 0.2.9]\n! Title: test\n! comment\n\n||example.com\n@@||good.cn\n/regex/\n"
	enc := base64.StdEncoding.EncodeToString([]byte(plain))
	// upstream wraps base64 in newlines; emulate that
	wrapped := ""
	for i := 0; i < len(enc); i += 10 {
		end := i + 10
		if end > len(enc) {
			end = len(enc)
		}
		wrapped += enc[i:end] + "\n"
	}
	f := writeFile(t, "gfwlist.txt", wrapped)

	rules, err := parseGFWList(f)
	if err != nil {
		t.Fatalf("parseGFWList: %v", err)
	}
	want := []string{"||example.com", "@@||good.cn", "/regex/"}
	if strings.Join(rules, "|") != strings.Join(want, "|") {
		t.Errorf("rules = %v, want %v", rules, want)
	}
}

func TestParseGFWList_Missing(t *testing.T) {
	rules, err := parseGFWList(filepath.Join(t.TempDir(), "nope.txt"))
	if err != nil {
		t.Fatalf("missing gfwlist should not error, got %v", err)
	}
	if rules != nil {
		t.Errorf("expected nil rules, got %v", rules)
	}
}

func TestGeneratePAC_Compose(t *testing.T) {
	domains := writeFile(t, "domains.txt", "mysite.com\n!directme.com\n")
	plain := "[AutoProxy 0.2.9]\n||gfwonly.com\n"
	gfw := writeFile(t, "gfwlist.txt", base64.StdEncoding.EncodeToString([]byte(plain)))

	js, err := GeneratePAC(domains, gfw)
	if err != nil {
		t.Fatalf("GeneratePAC: %v", err)
	}

	for _, want := range []string{
		`var userProxy = {"mysite.com":1};`,
		`var userDirect = {"directme.com":1};`,
		`"||gfwonly.com"`,
		"function FindProxyForURL",
		"function userDecision",
		"defaultMatcher",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("generated PAC missing %q", want)
		}
	}
}

func TestGeneratePAC_MissingDomains(t *testing.T) {
	if _, err := GeneratePAC(filepath.Join(t.TempDir(), "nope.txt"), ""); err == nil {
		t.Fatal("expected error for missing domains file")
	}
}

// userMatch mirrors the generated userDecision() so we can assert override
// semantics in Go.
func userMatch(proxy, direct map[string]bool, host string) string {
	host = strings.ToLower(host)
	p := strings.Split(host, ".")
	for i := 0; i < len(p)-1; i++ {
		s := strings.Join(p[i:], ".")
		if direct[s] {
			return "DIRECT"
		}
		if proxy[s] {
			return "PROXY"
		}
	}
	return ""
}

func TestUserOverrideSemantics(t *testing.T) {
	proxy := map[string]bool{"claude.ai": true, "example.com": true}
	direct := map[string]bool{"intranet.example.com": true}

	cases := []struct {
		host string
		want string
	}{
		{"claude.ai", "PROXY"},
		{"api.claude.ai", "PROXY"},
		{"example.com", "PROXY"},
		{"intranet.example.com", "DIRECT"},   // more specific direct wins over example.com proxy
		{"a.intranet.example.com", "DIRECT"}, // subdomain of a direct entry
		{"other.com", ""},                    // no user rule -> fall through to gfwlist
		{"notclaude.ai", ""},                 // suffix is "ai", not a rule
	}
	for _, c := range cases {
		if got := userMatch(proxy, direct, c.host); got != c.want {
			t.Errorf("userMatch(%q) = %q, want %q", c.host, got, c.want)
		}
	}
}
