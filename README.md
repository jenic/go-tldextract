# tldextract

`tldextract` is a small Go CLI that reads hostnames or URLs and prints the
unique registrable domains found in the input.

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

## Usage

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
go test ./cmd/tldextract -run '^$' -fuzz=FuzzExtractHostIgnoresDelimitedSuffixUserinfo -fuzztime=10s
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
