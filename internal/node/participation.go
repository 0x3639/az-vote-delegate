package node

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/0x3639/znn-sdk-go/api/embedded"
)

// Participation measures distinct retained requests with a recorded vote. The
// RPC does not provide historical pillar eligibility or deleted phase revisions,
// so this is deliberately not presented as a lifetime attendance score.
type Participation struct {
	Name  string `json:"name"`
	Voted int    `json:"voted"`
	Total int    `json:"total"`
}

func (n *Node) Participation(ctx context.Context) ([]Participation, error) {
	if _, err := n.chainTime(ctx); err != nil {
		return nil, err
	}
	pillars, err := n.Pillars(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := n.projects(ctx)
	if err != nil {
		return nil, err
	}
	ids, err := participationIDs(projects)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for name, p := range pillars {
		if p.RevokeTimestamp == 0 {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	result := make([]Participation, 0, len(names))
	for _, name := range names {
		votes, err := n.PillarVotes(ctx, name, ids)
		if err != nil {
			return nil, fmt.Errorf("cannot calculate participation for %s: %w", name, err)
		}
		p := Participation{Name: name, Total: len(ids)}
		for _, vote := range votes {
			if vote != nil {
				p.Voted++ // Yes, no and abstain each count as participation.
			}
		}
		result = append(result, p)
	}
	return result, nil
}

func participationIDs(projects []*embedded.Project) ([]string, error) {
	seen := map[string]bool{}
	for _, p := range projects {
		if p == nil {
			return nil, errors.New("nil project in participation history")
		}
		seen[p.Id.String()] = true
		phases := map[string]bool{}
		for _, phase := range p.Phases {
			if phase == nil || phase.Phase == nil || phase.Phase.ProjectID != p.Id {
				return nil, errors.New("invalid phase in participation history")
			}
			phases[phase.Phase.Id.String()] = true
		}
		for _, id := range p.PhaseIds {
			if !phases[id.String()] {
				return nil, errors.New("missing phase in participation history")
			}
			seen[id.String()] = true
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}
