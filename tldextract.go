// Package tldextract extracts registrable domains from URLs and host-like input.
package tldextract

import (
	"errors"
	"net"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

// ErrNoRegistrableDomain indicates that the input does not identify a
// registrable domain. This includes blank and comment input, IP addresses,
// local names, and invalid or unknown domains.
var ErrNoRegistrableDomain = errors.New("tldextract: no registrable domain")

// Options controls domain extraction output.
type Options struct {
	// Punycode returns domains in ASCII IDNA form. The default is Unicode.
	Punycode bool
}

// Extract returns the registrable domain in input according to the public
// suffix list. It accepts URLs, bare hosts, host:port values, bracketed IPv6
// literals, and userinfo-style authorities.
func Extract(input string, options Options) (string, error) {
	host := extractHost(input)
	if host == "" || strings.HasPrefix(host, "#") {
		return "", ErrNoRegistrableDomain
	}

	if net.ParseIP(host) != nil {
		return "", ErrNoRegistrableDomain
	}

	// Normalize Unicode hosts for public-suffix lookup when possible.
	if unicodeHost, err := idna.Lookup.ToUnicode(host); err == nil && unicodeHost != "" {
		host = unicodeHost
	}
	host = strings.TrimSuffix(strings.ToLower(host), ".")

	registrable, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil || registrable == "" {
		return "", ErrNoRegistrableDomain
	}

	if options.Punycode {
		if ascii, err := idna.Lookup.ToASCII(registrable); err == nil && ascii != "" {
			registrable = ascii
		}
	}

	return registrable, nil
}

// ExtractAll extracts unique registrable domains from inputs. Inputs that do
// not have a registrable domain are ignored, and returned domains are sorted.
func ExtractAll(inputs []string, options Options) ([]string, error) {
	seen := make(map[string]struct{})
	for _, input := range inputs {
		domain, err := Extract(input, options)
		if errors.Is(err, ErrNoRegistrableDomain) {
			continue
		}
		if err != nil {
			return nil, err
		}
		seen[domain] = struct{}{}
	}

	domains := make([]string, 0, len(seen))
	for domain := range seen {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	return domains, nil
}

func extractHost(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ".")

	// If it looks like a URL, parse normally.
	if hasScheme(s) {
		u, err := url.Parse(s)
		if err == nil && u.Hostname() != "" {
			return u.Hostname()
		}
		// Fall through to manual cleanup if URL parsing fails.
	}

	// Manual cleanup for bare host, host:port, [v6]:port, creds@host.
	// Cut at the first path, query, or fragment delimiter.
	for i, ch := range s {
		if ch == '/' || ch == '?' || ch == '#' {
			s = s[:i]
			break
		}
	}
	if at := strings.LastIndex(s, "@"); at != -1 {
		s = s[at+1:]
	}
	if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
		right := strings.IndexByte(s, ']')
		if right != -1 {
			return strings.Trim(s[1:right], " \t")
		}
	}
	if colon := strings.LastIndexByte(s, ':'); colon != -1 && !strings.Contains(s, "]") {
		s = s[:colon]
	}
	return s
}

func hasScheme(s string) bool {
	// Minimal check for schemes such as http:// and SSH+git://.
	i := strings.Index(s, "://")
	if i <= 0 {
		return false
	}

	for j, r := range s[:i] {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || (j > 0 && (r >= '0' && r <= '9' || r == '+' || r == '-' || r == '.'))) {
			return false
		}
	}
	return true
}
