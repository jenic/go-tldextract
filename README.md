# tldextract

`tldextract` is a Go library and CLI for extracting registrable domains from
hostnames or URLs.

It uses the public suffix list from `golang.org/x/net/publicsuffix`, so inputs
such as `www.example.co.uk` resolve to `example.co.uk` instead of just
`co.uk`.

## Features

- Reads newline-delimited input from stdin or a file.
- Accepts full URLs, bare hosts, `host:port`, bracketed IPv6 literals, and
  userinfo-style authorities.
- Skips empty lines, comment lines beginning with `#`, IP addresses, and names
  that do not have a public suffix match, such as `localhost`.
- Normalizes domains to lowercase.
- De-duplicates and sorts output.
- Supports Unicode domains and optional ASCII/punycode output.

## Requirements

- Go 1.25.11 or newer.

## Library Usage

Install the module in a Go project:

```sh
go get github.com/jenic/go-tldextract
```

Extract one domain and explicitly handle input without a registrable domain:

```go
import (
	"errors"
	"fmt"

	tldextract "github.com/jenic/go-tldextract"
)

domain, err := tldextract.Extract("https://www.example.co.uk/path", tldextract.Options{})
if err != nil {
	if errors.Is(err, tldextract.ErrNoRegistrableDomain) {
		// The input was an IP address, local name, comment, or invalid domain.
		return
	}
	panic(err)
}
fmt.Println(domain) // example.co.uk
```

Use `Options{Punycode: true}` for ASCII IDNA output. `ExtractAll` accepts a
slice of inputs, ignores values returning `ErrNoRegistrableDomain`, and returns
unique domains in sorted order:

```go
domains, err := tldextract.ExtractAll(inputs, tldextract.Options{Punycode: true})
if err != nil {
	panic(err)
}
```

## CLI Usage

Run from stdin:

```sh
printf '%s\n' \
  'https://www.example.com/path' \
  'subdomain.example.co.uk:443' \
  'localhost' \
  '192.0.2.1' |
go run ./cmd/tldextract
```

Output:

```text
example.co.uk
example.com
```

Run from a file:

```sh
go run ./cmd/tldextract input.txt
```

Use `-` explicitly for stdin:

```sh
go run ./cmd/tldextract - < input.txt
```

Output registrable domains as ASCII IDNA/punycode:

```sh
go run ./cmd/tldextract -punycode input.txt
```

## Input Format

The input is newline-delimited. Each non-empty, non-comment line is treated as a
candidate URL or hostname.

Examples of accepted input:

```text
https://user:pass@www.example.com:443/path
www.example.co.uk
example.com/path?ref=test
[2001:db8::1]:443
# ignored comment
```

Only registrable domains are printed. IP addresses and unknown local names are
ignored.

## Building

```sh
go build -o tldextract ./cmd/tldextract
```

Then run:

```sh
./tldextract input.txt
```

## Testing

Run the unit tests:

```sh
go test ./...
```

Run the fuzz target for host parsing:

```sh
go test . -run '^$' -fuzz=FuzzExtractHostIgnoresDelimitedSuffixUserinfo -fuzztime=10s
```

The fuzz target covers cases where path, query, or fragment text contains
userinfo-like `@` characters that must not replace the original host.

## Security Checks

Recommended checks for this project:

```sh
go mod verify
go vet ./...
govulncheck ./...
gosec ./...
```

When running in constrained environments, set `GOCACHE`, `GOMODCACHE`, and
`GOPATH` to workspace-local directories so Go does not write outside the
workspace.
