package pac

import (
	"regexp"
	"strings"
)

// LookupResult describes how a host would be routed by the current rules.
type LookupResult struct {
	Host    string
	Proxied bool
	Source  string // "domains.txt", "gfwlist", "gfwlist-whitelist", "default"
	Rule    string // the matched rule (empty for default)
}

// Lookup evaluates domains.txt (highest priority) then gfwlist.txt for host,
// mirroring the generated PAC's decision order.
func Lookup(domainsFile, gfwlistFile, host string) (LookupResult, error) {
	host = NormalizeDomain(host)
	res := LookupResult{Host: host}

	proxy, direct, err := parseDomains(domainsFile)
	if err != nil {
		return res, err
	}
	if src, ok, proxied := userLookup(proxy, direct, host); ok {
		res.Proxied = proxied
		res.Source = "domains.txt"
		res.Rule = src
		return res, nil
	}

	rules, err := parseGFWList(gfwlistFile)
	if err != nil {
		return res, err
	}
	proxied, whitelisted, rule := gfwlistDecision(rules, host)
	switch {
	case whitelisted:
		res.Proxied = false
		res.Source = "gfwlist-whitelist"
		res.Rule = rule
	case proxied:
		res.Proxied = true
		res.Source = "gfwlist"
		res.Rule = rule
	default:
		res.Proxied = false
		res.Source = "default"
	}
	return res, nil
}

// userLookup returns the matched rule text, whether a rule matched, and the
// decision. Most-specific subdomain wins; within a level, direct beats proxy.
func userLookup(proxy, direct []string, host string) (rule string, ok, proxied bool) {
	d := make(map[string]bool, len(direct))
	for _, x := range direct {
		d[x] = true
	}
	p := make(map[string]bool, len(proxy))
	for _, x := range proxy {
		p[x] = true
	}
	parts := strings.Split(host, ".")
	for i := 0; i < len(parts)-1; i++ {
		s := strings.Join(parts[i:], ".")
		if d[s] {
			return "!" + s, true, false
		}
		if p[s] {
			return s, true, true
		}
	}
	return "", false, false
}

// gfwlistDecision applies AutoProxy rules to test URLs for host. Whitelist (@@)
// rules override blocking rules, matching the Adblock matcher in engine.js.
func gfwlistDecision(rules []string, host string) (proxied, whitelisted bool, rule string) {
	urls := []string{"https://" + host + "/", "http://" + host + "/"}

	var blockRule, wlRule string
	for _, r := range rules {
		pat := r
		isException := false
		if strings.HasPrefix(pat, "@@") {
			isException = true
			pat = pat[2:]
		}
		matched := false
		for _, u := range urls {
			if matchAutoProxy(pat, u, host) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		if isException {
			if wlRule == "" {
				wlRule = r
			}
		} else if blockRule == "" {
			blockRule = r
		}
	}

	if wlRule != "" {
		return false, true, wlRule
	}
	if blockRule != "" {
		return true, false, blockRule
	}
	return false, false, ""
}

func matchAutoProxy(pat, url, host string) bool {
	if pat == "" {
		return false
	}
	// regex rule: /.../
	if len(pat) >= 2 && strings.HasPrefix(pat, "/") && strings.HasSuffix(pat, "/") {
		re, err := regexp.Compile(pat[1 : len(pat)-1])
		if err != nil {
			return false
		}
		return re.MatchString(url)
	}
	// domain anchor: ||domain
	if strings.HasPrefix(pat, "||") {
		return matchDomainAnchor(pat[2:], url, host)
	}
	// start anchor: |http://...
	if strings.HasPrefix(pat, "|") {
		return wildMatch(pat[1:], url, true)
	}
	// keyword / substring (a leading '.' is literal)
	return wildMatch(pat, url, false)
}

func matchDomainAnchor(d, url, host string) bool {
	hostPat := d
	pathPat := ""
	if i := strings.IndexByte(d, '/'); i >= 0 {
		hostPat = d[:i]
		pathPat = d[i:]
	}
	if !domainBoundaryMatch(hostPat, host) {
		return false
	}
	if pathPat == "" || pathPat == "/" {
		return true
	}
	return strings.Contains(url, pathPat)
}

func domainBoundaryMatch(hostPat, host string) bool {
	if strings.ContainsAny(hostPat, "*^") {
		re, err := regexp.Compile(`(^|\.)` + wildcardToRegex(hostPat) + `$`)
		if err != nil {
			return false
		}
		return re.MatchString(host)
	}
	return host == hostPat || strings.HasSuffix(host, "."+hostPat)
}

func wildMatch(pat, url string, anchorStart bool) bool {
	anchorEnd := false
	if strings.HasSuffix(pat, "|") {
		anchorEnd = true
		pat = pat[:len(pat)-1]
	}
	if strings.ContainsAny(pat, "*^") {
		re := wildcardToRegex(pat)
		if anchorStart {
			re = "^" + re
		}
		if anchorEnd {
			re = re + "$"
		}
		r, err := regexp.Compile(re)
		if err != nil {
			return false
		}
		return r.MatchString(url)
	}
	switch {
	case anchorStart && anchorEnd:
		return url == pat
	case anchorStart:
		return strings.HasPrefix(url, pat)
	case anchorEnd:
		return strings.HasSuffix(url, pat)
	default:
		return strings.Contains(url, pat)
	}
}

// wildcardToRegex converts AutoProxy wildcard syntax to a Go regexp fragment:
// `*` => any run, `^` => a separator char or end-of-string.
func wildcardToRegex(p string) string {
	var b strings.Builder
	for _, c := range p {
		switch c {
		case '*':
			b.WriteString(".*")
		case '^':
			b.WriteString(`(?:[/:?=&]|$)`)
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	return b.String()
}
