package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	for _, tc := range []struct {
		args     []string
		bad      bool
		contains string
	}{
		{nil, true, ""}, {[]string{"unknown"}, true, ""}, {[]string{"version"}, false, "az-vote-delegate"},
		{[]string{"version", "extra"}, true, ""}, {[]string{"run", "--help"}, false, ""},
		{[]string{"run", "--reset-state"}, true, ""}, {[]string{"status", "--dry-run=false"}, true, ""},
		{[]string{"run", "--password", "secret"}, true, ""},
	} {
		var out, stderr bytes.Buffer
		err := execute(context.Background(), tc.args, &out, &stderr)
		if (err != nil) != tc.bad {
			t.Fatalf("%v: %v", tc.args, err)
		}
		if !strings.Contains(out.String(), tc.contains) {
			t.Fatal(out.String())
		}
	}
}
