package tui

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/0x3639/az-vote-delegate/internal/control"
	"github.com/0x3639/az-vote-delegate/internal/delegate"
	"github.com/0x3639/az-vote-delegate/internal/node"
)

func TestTopThirtyExcludesSelfAndShufflesAfterRanking(t *testing.T) {
	all := []node.Participation{{Name: "self", Voted: 1000, Total: 1000}}
	for i := range 40 {
		all = append(all, node.Participation{Name: fmt.Sprintf("pillar-%02d", i), Voted: i, Total: 50})
	}
	original := append([]node.Participation(nil), all...)
	// A deterministic shuffle lets us assert the displayed order is shuffled,
	// rather than merely testing the ranking helper against its own comparator.
	reverse := func(n int, swap func(int, int)) {
		for i := 0; i < n/2; i++ {
			swap(i, n-1-i)
		}
	}
	got := topCandidates(all, "self", reverse)
	if len(got) != 30 {
		t.Fatal(len(got))
	}
	for i, p := range got {
		if p.Voted != i+10 || p.Name == "self" || p.Total != 50 {
			t.Fatalf("wrong top-30 membership/order/participation at %d: %+v", i, p)
		}
	}
	if !reflect.DeepEqual(all, original) {
		t.Fatal("picker mutated service data")
	}
	if got := topCandidates(all[:1], "self", reverse); len(got) != 0 {
		t.Fatal("self was selectable")
	}
}

func TestTieAtThirtyIsShuffledBeforeCutoff(t *testing.T) {
	all := make([]node.Participation, 31)
	for i := range all {
		all[i] = node.Participation{Name: fmt.Sprint(i), Voted: 1, Total: 1}
	}
	calls := 0
	got := topCandidates(all, "self", func(n int, swap func(int, int)) {
		calls++
		swap(0, n-1)
	})
	seen := map[string]bool{}
	for _, p := range got {
		seen[p.Name] = true
	}
	if calls != 2 || !seen["30"] || seen["0"] {
		t.Fatal("equal-vote cutoff was biased by input order", got)
	}
}

func TestUnansweredListsExcludeAllVoteValuesAndQueuedManualChoices(t *testing.T) {
	s := &control.Snapshot{State: delegate.NewState(1)}
	for _, kind := range []string{"project", "phase"} {
		for _, vote := range []*delegate.Vote{nil, ptr(delegate.Yes), ptr(delegate.No), ptr(delegate.Abstain)} {
			s.Rows = append(s.Rows, delegate.Row{Pillar: "A", Votable: delegate.Votable{ID: fmt.Sprint(len(s.Rows)), Kind: kind}, Owner: vote})
		}
	}
	if got := unvotedRows(s, false); !reflect.DeepEqual(got, []int{0}) {
		t.Fatal("project list", got)
	}
	if got := unvotedRows(s, true); !reflect.DeepEqual(got, []int{4}) {
		t.Fatal("phase list", got)
	}
	s.State.Entries("A")["4"] = &delegate.Entry{ManualVote: ptr(delegate.No)}
	if len(unvotedRows(s, true)) != 0 {
		t.Fatal("queued manual vote shown again")
	}
	s.State.Entries("A")["0"] = &delegate.Entry{Pending: &delegate.Pending{Vote: delegate.Yes}}
	if len(unvotedRows(s, false)) != 1 {
		t.Fatal("pending copy cannot be overridden before it executes")
	}
}
func ptr(v delegate.Vote) *delegate.Vote { return &v }
