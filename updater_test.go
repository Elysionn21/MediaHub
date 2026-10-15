package main

import (
	"strings"
	"testing"
)

func TestParseSHA256Digest(t *testing.T) {
	const digest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	got, err := parseSHA256Digest("sha256:" + strings.ToUpper(digest))
	if err != nil {
		t.Fatal(err)
	}
	if got != digest {
		t.Fatalf("got %q, want %q", got, digest)
	}
	for _, invalid := range []string{"", digest, "sha256:xyz", "sha256:" + digest[:63]} {
		if _, err := parseSHA256Digest(invalid); err == nil {
			t.Errorf("parseSHA256Digest(%q) unexpectedly succeeded", invalid)
		}
	}
}

func TestNewerVersion(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v0.1.1", "v0.1.0", true},
		{"v0.1.0", "v0.1.0", false},
		{"v0.1.0-rc.1", "v0.1.0", false},
		{"v0.2.0-beta.2", "v0.2.0-beta.1", true},
		{"not-a-version", "v0.1.0", false},
	}
	for _, tc := range cases {
		if got := newerVersion(tc.latest, tc.current); got != tc.want {
			t.Errorf("newerVersion(%q, %q) = %t, want %t", tc.latest, tc.current, got, tc.want)
		}
	}
}
