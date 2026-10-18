package tui

import (
	"strings"
	"testing"
)

func TestCleanChainText(t *testing.T) {
	got := Clean("Phase\x1b[2J\n\r\t\x07\u202e payout")
	if strings.ContainsAny(got, "\x1b\n\r\t\x07\u202e") {
		t.Fatal("terminal controls retained")
	}
}
