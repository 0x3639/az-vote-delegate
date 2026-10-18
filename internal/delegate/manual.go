package delegate

import (
	"context"
	"errors"
	"time"

	"github.com/0x3639/az-vote-delegate/internal/config"
)

// QueueManual is called by the single state owner between reconciliations.
// Persistence precedes acknowledgement. Existing pending sends are retained;
// the manual request waits for their receipt before it can be submitted.
func (e *Engine) QueueManual(ctx context.Context, d config.Delegation, id string, v Vote) error {
	if e.DryRun {
		return errors.New("enable live voting before casting a manual vote")
	}
	if !v.Valid() {
		return errors.New("invalid vote")
	}
	vs, err := e.Backend.ListVotables(ctx)
	if err != nil {
		return err
	}
	found := false
	for _, item := range vs {
		if item.ID == id {
			found = true
			break
		}
	}
	if !found {
		return errors.New("this project or payment request is no longer open")
	}
	entry := e.State.Entries(d.PillarName)[id]
	if entry == nil {
		entry = &Entry{}
		e.State.Entries(d.PillarName)[id] = entry
	}
	entry.CopyObservation = nil
	entry.ManualVote = &v
	entry.ManualConfirmed = false
	entry.OwnerOverride = true
	entry.Failures = 0
	entry.RetryAt = time.Time{}
	return e.persist()
}

func (e *Engine) Resume(ctx context.Context, d config.Delegation, id string) error {
	entry := e.State.Entries(d.PillarName)[id]
	if entry == nil {
		return nil
	}
	if entry.Pending != nil {
		return errors.New("wait for the pending transaction to finish before resuming copy voting")
	}
	votes, err := e.Backend.PillarVotes(ctx, d.PillarName, []string{id})
	if err != nil {
		return err
	}
	// The operator explicitly hands this round back to the mirror, so adopt
	// its current vote as the baseline rather than rediscovering an override.
	entry.Cast = votes[id]
	entry.CopyObservation = nil
	entry.ManualVote = nil
	entry.ManualConfirmed = false
	entry.OwnerOverride = false
	entry.Failures = 0
	entry.RetryAt = time.Time{}
	return e.persist()
}
