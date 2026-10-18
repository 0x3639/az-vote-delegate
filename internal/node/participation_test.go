package node

import (
	"context"
	"reflect"
	"testing"

	"github.com/0x3639/az-vote-delegate/internal/delegate"
	"github.com/0x3639/znn-sdk-go/api/embedded"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

func TestParticipationCountsRetainedProjectsAndPhases(t *testing.T) {
	n, f := newNode(t)
	a, b, c, obsolete := hash("a"), hash("b"), hash("c"), hash("d")
	f.projects = []*embedded.Project{{Id: a, Status: 2, Votes: &embedded.VoteBreakdown{}, PhaseIds: []types.Hash{b, c}, Phases: []*embedded.Phase{
		{Phase: &embedded.PhaseInfo{Id: b, ProjectID: a, Status: 1}},
		{Phase: &embedded.PhaseInfo{Id: c, ProjectID: a, Status: 0}},
	}}}
	yes, no, abstain := delegate.Yes, delegate.No, delegate.Abstain
	f.pillarVotes = map[string]delegate.Votes{
		"A": {a.String(): &yes, b.String(): &no, c.String(): &abstain},
		"B": {b.String(): &no, obsolete.String(): &yes},
	}
	f.pillarList = []map[string]any{
		{"name": "A", "weight": "0", "currentStats": map[string]any{}},
		{"name": "B", "weight": "0", "currentStats": map[string]any{}},
		{"name": "Revoked", "weight": "0", "currentStats": map[string]any{}, "revokeTimestamp": 1},
	}
	got, err := n.Participation(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []Participation{{Name: "A", Voted: 3, Total: 3}, {Name: "B", Voted: 1, Total: 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("participation = %+v, want %+v", got, want)
	}
	if !reflect.DeepEqual(f.voteSizes, []int{3, 3}) {
		t.Fatalf("queried obsolete IDs or revoked pillars: %v", f.voteSizes)
	}
	// No requests is unknown participation, not a fabricated 100%.
	f.projects = nil
	got, err = n.Participation(context.Background())
	if err != nil || got[0].Total != 0 || got[0].Voted != 0 {
		t.Fatalf("empty history: %+v, %v", got, err)
	}
}

func TestParticipationRejectsIncompleteHistoryOrVotes(t *testing.T) {
	n, f := newNode(t)
	a, b := hash("a"), hash("b")
	f.projects = []*embedded.Project{{Id: a, Votes: &embedded.VoteBreakdown{}, PhaseIds: []types.Hash{b}}}
	if _, err := n.Participation(context.Background()); err == nil {
		t.Fatal("missing phase produced a misleading participation rate")
	}
	f.projects[0].PhaseIds = nil
	f.votes = []*definition.PillarVote{}
	if _, err := n.Participation(context.Background()); err == nil {
		t.Fatal("incomplete vote response accepted")
	}
}
