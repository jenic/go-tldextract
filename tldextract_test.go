package tldextract

import (
	"errors"
	"reflect"
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		options Options
		want    string
		wantErr error
	}{
		{name: "URL", input: "https://user:pass@www.example.co.uk:443/path", want: "example.co.uk"},
		{name: "bare host with port", input: "www.example.com:443", want: "example.com"},
		{name: "Unicode domain", input: "www.b\u00fccher.de", want: "b\u00fccher.de"},
		{name: "punycode domain", input: "www.b\u00fccher.de", options: Options{Punycode: true}, want: "xn--bcher-kva.de"},
		{name: "comment", input: "# ignored", wantErr: ErrNoRegistrableDomain},
		{name: "IP address", input: "192.0.2.1", wantErr: ErrNoRegistrableDomain},
		{name: "localhost", input: "localhost", wantErr: ErrNoRegistrableDomain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Extract(tt.input, tt.options)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Extract(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("Extract(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractAll(t *testing.T) {
	got, err := ExtractAll([]string{
		"https://www.example.com/path",
		"www.example.co.uk:443",
		"example.com",
		"localhost",
		"192.0.2.1",
		"# ignored",
	}, Options{})
	if err != nil {
		t.Fatalf("ExtractAll() error = %v", err)
	}

	want := []string{"example.co.uk", "example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractAll() = %v, want %v", got, want)
	}
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "bare host", in: "example.com", want: "example.com"},
		{name: "bare host with trailing dot", in: "example.com.", want: "example.com"},
		{name: "bare host with port", in: "example.com:443", want: "example.com"},
		{name: "bracketed IPv6 with port", in: "[2001:db8::1]:443", want: "2001:db8::1"},
		{name: "bare userinfo", in: "user:pass@example.com:443", want: "example.com"},
		{name: "URL userinfo", in: "https://user:pass@example.com:443/path@evil.com", want: "example.com"},
		{name: "path at sign does not replace host", in: "example.com/path@evil.com", want: "example.com"},
		{name: "query at sign does not replace host", in: "example.com?x=@evil.com", want: "example.com"},
		{name: "fragment at sign does not replace host", in: "example.com#@evil.com", want: "example.com"},
		{name: "path query at sign does not replace host", in: "example.com/path?x=@evil.com", want: "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractHost(tt.in); got != tt.want {
				t.Fatalf("extractHost(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func FuzzExtractHostIgnoresDelimitedSuffixUserinfo(f *testing.F) {
	for _, seed := range []string{
		"@evil.com",
		"x=@evil.com",
		"user:pass@evil.com",
		"%40evil.com",
		"nested/path@evil.com",
		"query?x=@evil.com",
		"fragment#@evil.com",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, suffix string) {
		for _, delimiter := range []string{"/", "?", "#"} {
			input := "example.com" + delimiter + suffix
			if got := extractHost(input); got != "example.com" {
				t.Fatalf("extractHost(%q) = %q, want example.com", input, got)
			}
		}
	})
}
