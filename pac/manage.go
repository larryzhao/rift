package pac

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// NormalizeDomain lowercases and strips scheme/path/port so callers can pass a
// bare domain or a pasted URL.
func NormalizeDomain(input string) string {
	d := strings.TrimSpace(strings.ToLower(input))
	if i := strings.Index(d, "://"); i >= 0 {
		d = d[i+3:]
	}
	if i := strings.IndexAny(d, "/?#"); i >= 0 {
		d = d[:i]
	}
	if i := strings.IndexByte(d, ':'); i >= 0 {
		d = d[:i]
	}
	d = strings.TrimPrefix(d, "!")
	return strings.Trim(d, ".")
}

// SetDomainRule adds or updates a domain in domainsFile. direct=true writes
// "!domain" (force DIRECT); direct=false writes "domain" (force proxy). Any
// existing entry for the same domain (either form) is replaced. Comments and
// other entries are preserved. The file is created if it does not exist.
func SetDomainRule(domainsFile, domain string, direct bool) error {
	domain = NormalizeDomain(domain)
	if domain == "" {
		return fmt.Errorf("empty domain")
	}

	var lines []string
	if f, err := os.Open(domainsFile); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			trimmed := strings.TrimSpace(line)
			// drop existing entry for this domain (either proxy or direct form)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				if NormalizeDomain(trimmed) == domain {
					continue
				}
			}
			lines = append(lines, line)
		}
		f.Close()
		if err := sc.Err(); err != nil {
			return fmt.Errorf("read domains file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("open domains file: %w", err)
	}

	entry := domain
	if direct {
		entry = "!" + domain
	}
	lines = append(lines, entry)

	out := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(domainsFile, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write domains file: %w", err)
	}
	return nil
}
