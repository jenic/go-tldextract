package main

import "testing"

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "bare host",
			in:   "example.com",
			want: "example.com",
		},
		{
			name: "bare host with trailing dot",
			in:   "example.com.",
			want: "example.com",
		},
		{
			name: "bare host with port",
			in:   "example.com:443",
			want: "example.com",
		},
		{
			name: "bracketed ipv6 with port",
			in:   "[2001:db8::1]:443",
			want: "2001:db8::1",
		},
		{
			name: "bare userinfo",
			in:   "user:pass@example.com:443",
			want: "example.com",
		},
		{
			name: "url userinfo",
			in:   "https://user:pass@example.com:443/path@evil.com",
			want: "example.com",
		},
		{
			name: "path at sign does not replace host",
			in:   "example.com/path@evil.com",
			want: "example.com",
		},
		{
			name: "query at sign does not replace host",
			in:   "example.com?x=@evil.com",
			want: "example.com",
		},
		{
			name: "fragment at sign does not replace host",
			in:   "example.com#@evil.com",
			want: "example.com",
		},
		{
			name: "path query at sign does not replace host",
			in:   "example.com/path?x=@evil.com",
			want: "example.com",
		},
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
