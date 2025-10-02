package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

func main() {
	punycode := flag.Bool("punycode", false, "output registrable domains in ASCII (IDNA/punycode)")
	flag.Parse()

	in := os.Stdin
	if flag.NArg() > 0 && flag.Arg(0) != "-" {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "open: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	seen := make(map[string]struct{})
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		host := extractHost(line)
		if host == "" {
			continue
		}
		// Skip IPs
		if ip := net.ParseIP(host); ip != nil {
			continue
		}

		// Normalize Unicode host to a canonical form for PSL lookup.
		uHost, err := idna.Lookup.ToUnicode(host)
		if err == nil && uHost != "" {
			host = uHost
		}
		host = strings.TrimSuffix(strings.ToLower(host), ".")

		reg, err := publicsuffix.EffectiveTLDPlusOne(host)
		if err != nil || reg == "" {
			// If PSL has no idea (e.g., "localhost"), skip.
			continue
		}

		if *punycode {
			if a, err := idna.Lookup.ToASCII(reg); err == nil && a != "" {
				reg = a
			}
		}
		seen[reg] = struct{}{}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}

	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}
	sort.Strings(out)
	for _, d := range out {
		fmt.Println(d)
	}
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
		// fall through to manual cleanup if URL parse fails
	}

	// Manual cleanup for bare host, host:port, [v6]:port, creds@host…
	// Strip creds
	if at := strings.LastIndex(s, "@"); at != -1 {
		s = s[at+1:]
	}
	// Cut at first / ? #
	for i, ch := range s {
		if ch == '/' || ch == '?' || ch == '#' {
			s = s[:i]
			break
		}
	}
	// Strip :port (keeps IPv6 literals in [::1])
	if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
		// [IPv6]:port -> inside brackets
		right := strings.IndexByte(s, ']')
		if right != -1 {
			return strings.Trim(s[1:right], " \t")
		}
	}
	// host:port -> host
	if colon := strings.LastIndexByte(s, ':'); colon != -1 && !strings.Contains(s, "]") {
		s = s[:colon]
	}
	return s
}

func hasScheme(s string) bool {
	// minimal check: something like "http://" or "SSH+git://"
	i := strings.Index(s, "://")
	if i <= 0 {
		return false
	}
	// ensure scheme is a valid start char and chars
	sch := s[:i]
	for j, r := range sch {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || (j > 0 && (r >= '0' && r <= '9' || r == '+' || r == '-' || r == '.'))) {
			return false
		}
	}
	return true
}
