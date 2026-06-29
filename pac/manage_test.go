package pac

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

func TestNormalizeDomain(t *testing.T) {
	cases := map[string]string{
		"Claude.AI":                 "claude.ai",
		"  github.com  ":            "github.com",
		"https://api.openai.com/v1": "api.openai.com",
		"http://x.com:8080/path":    "x.com",
		"!twitter.com":              "twitter.com",
		".example.com.":             "example.com",
	}
	for in, want := range cases {
		if got := NormalizeDomain(in); got != want {
			t.Errorf("NormalizeDomain(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSetDomainRule(t *testing.T) {
	f := writeFile(t, "domains.txt", "# header\nexisting.com\n")

	if err := SetDomainRule(f, "new.com", false); err != nil {
		t.Fatal(err)
	}
	if err := SetDomainRule(f, "blocked.com", true); err != nil {
		t.Fatal(err)
	}
	// updating an existing domain replaces (not duplicates) it
	if err := SetDomainRule(f, "new.com", true); err != nil {
		t.Fatal(err)
	}

	b, _ := os.ReadFile(f)
	content := string(b)

	if !strings.Contains(content, "# header") {
		t.Error("comment header not preserved")
	}
	if !strings.Contains(content, "existing.com") {
		t.Error("existing entry not preserved")
	}
	if !strings.Contains(content, "!blocked.com") {
		t.Error("direct entry not written")
	}
	if strings.Count(content, "new.com") != 1 {
		t.Errorf("new.com should appear once, content:\n%s", content)
	}
	if !strings.Contains(content, "!new.com") {
		t.Error("new.com should have been flipped to direct")
	}
}

func TestSetDomainRule_CreatesFile(t *testing.T) {
	f := writeFile(t, "domains.txt", "")
	os.Remove(f)
	if err := SetDomainRule(f, "fresh.com", false); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(f)
	if !strings.Contains(string(b), "fresh.com") {
		t.Errorf("expected fresh.com, got %q", b)
	}
}

func gfwFile(t *testing.T, lines ...string) string {
	t.Helper()
	plain := "[AutoProxy 0.2.9]\n" + strings.Join(lines, "\n") + "\n"
	return writeFile(t, "gfwlist.txt", base64.StdEncoding.EncodeToString([]byte(plain)))
}

func TestLookup(t *testing.T) {
	domains := writeFile(t, "domains.txt", "mysite.com\n!twitter.com\n")
	gfw := gfwFile(t, "||twitter.com", "||example.com", "@@||good.example.com")

	cases := []struct {
		host        string
		wantProxied bool
		wantSource  string
	}{
		{"mysite.com", true, "domains.txt"},              // user proxy
		{"api.mysite.com", true, "domains.txt"},          // subdomain
		{"twitter.com", false, "domains.txt"},            // user override beats gfwlist
		{"example.com", true, "gfwlist"},                 // gfwlist block
		{"sub.example.com", true, "gfwlist"},             // gfwlist subdomain
		{"good.example.com", false, "gfwlist-whitelist"}, // gfwlist whitelist
		{"baidu.com", false, "default"},                  // nothing matches
	}
	for _, c := range cases {
		res, err := Lookup(domains, gfw, c.host)
		if err != nil {
			t.Fatalf("Lookup(%q): %v", c.host, err)
		}
		if res.Proxied != c.wantProxied || res.Source != c.wantSource {
			t.Errorf("Lookup(%q) = {proxied:%v source:%q}, want {proxied:%v source:%q}",
				c.host, res.Proxied, res.Source, c.wantProxied, c.wantSource)
		}
	}
}
