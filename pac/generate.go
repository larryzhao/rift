package pac

import (
	"bufio"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// engine is the Adblock-Plus filter matcher (from gfwlist2pac), embedded so the
// generated PAC can evaluate raw AutoProxy/gfwlist rules natively. It ends with
// `var defaultMatcher = new CombinedMatcher();`.
//
//go:embed engine.js
var engine string

const proxyLine = `SOCKS5 127.0.0.1:6153; SOCKS 127.0.0.1:6153; DIRECT;`

// GeneratePAC compiles a PAC script from two sources:
//
//   - domainsFile: the user's own list, one domain per line. A bare domain is
//     forced through the proxy; a `!`-prefixed domain is forced DIRECT. These
//     always override gfwlist. `#` starts a comment; blank lines are ignored.
//   - gfwlistFile: a base64-encoded gfwlist (AutoProxy format) synced from
//     upstream. Optional — if it is empty/missing, only the user rules apply.
//
// Decision order at runtime: user rules first (most-specific subdomain wins),
// then gfwlist, else DIRECT.
func GeneratePAC(domainsFile, gfwlistFile string) (string, error) {
	userProxy, userDirect, err := parseDomains(domainsFile)
	if err != nil {
		return "", err
	}

	rules, err := parseGFWList(gfwlistFile)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "var proxy = %q;\n", proxyLine)
	fmt.Fprintf(&b, "var userProxy = {%s};\n", dictEntries(userProxy))
	fmt.Fprintf(&b, "var userDirect = {%s};\n", dictEntries(userDirect))
	fmt.Fprintf(&b, "var rules = [%s];\n", ruleEntries(rules))
	b.WriteString(engine)
	b.WriteString(`
for (var i = 0; i < rules.length; i++) {
  defaultMatcher.add(Filter.fromText(rules[i]));
}

function userDecision(host) {
  host = host.toLowerCase();
  var p = host.split(".");
  for (var i = 0; i < p.length - 1; i++) {
    var s = p.slice(i).join(".");
    if (userDirect[s]) return "DIRECT";
    if (userProxy[s]) return proxy;
  }
  return "";
}

function FindProxyForURL(url, host) {
  var u = userDecision(host);
  if (u) return u;
  if (defaultMatcher.matchesAny(url, host) instanceof BlockingFilter) {
    return proxy;
  }
  return "DIRECT";
}
`)
	return b.String(), nil
}

// parseDomains reads the user domain list into proxy and direct sets.
func parseDomains(domainsFile string) (proxy, direct []string, err error) {
	f, err := os.Open(domainsFile)
	if err != nil {
		return nil, nil, fmt.Errorf("open domains file: %w", err)
	}
	defer f.Close()

	seen := make(map[string]bool)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		toDirect := false
		if strings.HasPrefix(line, "!") {
			toDirect = true
			line = strings.TrimSpace(strings.TrimPrefix(line, "!"))
		}
		if line == "" {
			continue
		}
		d := strings.ToLower(line)
		if seen[d] {
			continue
		}
		seen[d] = true
		if toDirect {
			direct = append(direct, d)
		} else {
			proxy = append(proxy, d)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, nil, fmt.Errorf("read domains file: %w", err)
	}
	return proxy, direct, nil
}

// parseGFWList decodes a base64 gfwlist file into a slice of AutoProxy rule
// lines (comments, the [AutoProxy] header, and blank lines stripped). A missing
// file is treated as an empty rule set so the PAC still works on user rules.
func parseGFWList(gfwlistFile string) ([]string, error) {
	if gfwlistFile == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(gfwlistFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read gfwlist file: %w", err)
	}

	// gfwlist.txt is base64 with embedded newlines; strip whitespace first.
	clean := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, string(raw))
	if clean == "" {
		return nil, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("base64 decode gfwlist: %w", err)
	}

	var rules []string
	sc := bufio.NewScanner(strings.NewReader(string(decoded)))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "[") {
			continue
		}
		rules = append(rules, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan gfwlist: %w", err)
	}
	return rules, nil
}

func dictEntries(domains []string) string {
	parts := make([]string, len(domains))
	for i, d := range domains {
		parts[i] = fmt.Sprintf("%q:1", d)
	}
	return strings.Join(parts, ",")
}

func ruleEntries(rules []string) string {
	parts := make([]string, len(rules))
	for i, r := range rules {
		parts[i] = fmt.Sprintf("%q", r)
	}
	return strings.Join(parts, ",")
}
