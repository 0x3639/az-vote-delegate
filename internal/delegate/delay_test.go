package delegate

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCopyDelaySurvivesRestartAndWaitsUntilDue(t *testing.T) {
	for _, kind := range []string{"project", "phase"} {
		t.Run(kind, func(t *testing.T) {
			e, f, ds := fixture()
			f.vs[0].Kind = kind
			now := e.now()
			e.Now = func() time.Time { return now }
			e.CopyDelay = 10 * time.Minute
			path := filepath.Join(t.TempDir(), "state.json")
			e.Save = func(s *State) error { return SaveState(path, s) }
			rows := reconcile(t, e, ds)
			if f.sends != 0 || rows[0].Action != "waiting-delay" || !rows[0].CopyAfter.Equal(now.Add(10*time.Minute)) {
				t.Fatal(rows, f.sends)
			}
			state, err := LoadState(path, 1)
			if err != nil {
				t.Fatal(err)
			}
			e.State = state // Simulate restarting from the persisted observation.
			now = now.Add(10*time.Minute - time.Second)
			reconcile(t, e, ds)
			if f.sends != 0 {
				t.Fatal("sent before deadline")
			}
			now = now.Add(time.Second)
			reconcile(t, e, ds)
			if f.sends != 1 || f.votes["A"][testID] == nil {
				t.Fatal("did not send at deadline")
			}
			reconcile(t, e, ds)
			if f.sends != 1 {
				t.Fatal("duplicate vote")
			}
		})
	}
}

func TestDelayRestartsWhenSourceChangesOrDisappears(t *testing.T) {
	for _, change := range []string{"vote", "pillar", "withdrawn", "reset"} {
		t.Run(change, func(t *testing.T) {
			e, f, ds := fixture()
			now := e.now()
			e.Now = func() time.Time { return now }
			e.CopyDelay = 10 * time.Minute
			reconcile(t, e, ds)
			now = now.Add(9 * time.Minute)
			switch change {
			case "vote":
				f.votes["B"][testID] = ptr(No)
			case "pillar":
				f.votes["C"] = Votes{testID: ptr(Yes)}
				ds[0].Follow = "C"
			case "withdrawn":
				f.votes["B"][testID] = nil
				reconcile(t, e, ds)
				f.votes["B"][testID] = ptr(Yes)
			case "reset":
				e.State.Entries("A")[testID].Cast = ptr(No)
			}
			rows := reconcile(t, e, ds)
			if !rows[0].CopyAfter.Equal(now.Add(10 * time.Minute)) {
				t.Fatal("old timer reused", rows)
			}
			now = now.Add(time.Minute)
			reconcile(t, e, ds)
			if f.sends != 0 {
				t.Fatal("sent at old deadline")
			}
			now = now.Add(9 * time.Minute)
			reconcile(t, e, ds)
			if f.sends != 1 {
				t.Fatal("did not send after new delay")
			}
		})
	}
}

func TestManualVoteBypassesDelayWithoutBlockingOtherRequests(t *testing.T) {
	e, f, ds := fixture()
	other := strings.Repeat("c", 64)
	f.vs = append(f.vs, Votable{ID: other, Kind: "phase"})
	f.vs[0].Kind = "project"
	f.votes["B"][other] = ptr(No)
	now := e.now()
	e.Now = func() time.Time { return now }
	e.CopyDelay = time.Hour
	reconcile(t, e, ds)
	if err := e.QueueManual(context.Background(), ds[0], testID, Abstain); err != nil {
		t.Fatal(err)
	}
	reconcile(t, e, ds)
	if f.sends != 1 || !Equal(f.votes["A"][testID], ptr(Abstain)) || f.votes["A"][other] != nil {
		t.Fatal("manual vote delayed or other request copied early")
	}
	now = now.Add(time.Hour)
	reconcile(t, e, ds)
	if f.sends != 2 || !Equal(f.votes["A"][other], ptr(No)) || !Equal(f.votes["A"][testID], ptr(Abstain)) {
		t.Fatal("manual override blocked another request or was overwritten")
	}
}

func TestDelayedVoteIsRecheckedBeforeSend(t *testing.T) {
	e, f, ds := fixture()
	now := e.now()
	e.Now = func() time.Time { return now }
	e.CopyDelay = time.Minute
	reconcile(t, e, ds)
	now = now.Add(time.Minute)
	f.during = func() { f.votes["B"][testID] = ptr(No) }
	rows := reconcile(t, e, ds)
	if rows[0].Action != "changed" || f.votes["A"][testID] != nil || e.State.Entries("A")[testID].CopyObservation != nil {
		t.Fatal("stale vote submitted", rows)
	}
	f.during = nil
	rows = reconcile(t, e, ds)
	if rows[0].Action != "waiting-delay" || !rows[0].CopyAfter.Equal(now.Add(time.Minute)) {
		t.Fatal("new vote did not get a new timer")
	}
}

func TestDelayClosedRequestAndExternalOverride(t *testing.T) {
	for _, closed := range []bool{false, true} {
		e, f, ds := fixture()
		now := e.now()
		e.Now = func() time.Time { return now }
		e.CopyDelay = time.Minute
		reconcile(t, e, ds)
		now = now.Add(time.Minute)
		if closed {
			f.vs = nil
		} else {
			f.votes["A"][testID] = ptr(Yes)
		}
		reconcile(t, e, ds)
		if f.sends != 0 {
			t.Fatal("copied a closed or manually voted request")
		}
		if entry := e.State.Entries("A")[testID]; entry != nil && entry.CopyObservation != nil {
			t.Fatal("stale timer retained")
		}
	}
}

func TestDelayChangesApplyToWaitingVotes(t *testing.T) {
	e, f, ds := fixture()
	now := e.now()
	e.Now = func() time.Time { return now }
	e.CopyDelay = time.Minute
	reconcile(t, e, ds)
	now = now.Add(time.Minute)
	e.CopyDelay = 2 * time.Minute
	reconcile(t, e, ds)
	if f.sends != 0 {
		t.Fatal("ignored increased delay")
	}
	e.CopyDelay = 0
	reconcile(t, e, ds)
	if f.sends != 1 {
		t.Fatal("disabling delay did not release vote")
	}
}

func TestDelayPersistenceFailurePreventsSend(t *testing.T) {
	e, f, ds := fixture()
	e.CopyDelay = time.Minute
	e.Save = func(*State) error { return errors.New("disk full") }
	_, err := e.Reconcile(context.Background(), ds)
	if !errors.Is(err, ErrPersistence) || f.sends != 0 {
		t.Fatal(err, f.sends)
	}
}

func TestDryRunDelayDoesNotWriteOrSubmit(t *testing.T) {
	e, f, ds := fixture()
	now := e.now()
	e.Now = func() time.Time { return now }
	e.CopyDelay = time.Minute
	e.DryRun = true
	e.Save = func(*State) error { t.Fatal("dry run wrote state"); return nil }
	reconcile(t, e, ds)
	now = now.Add(time.Minute)
	rows := reconcile(t, e, ds)
	if f.sends != 0 || rows[0].Action != "would-mirror" {
		t.Fatal(rows, f.sends)
	}
}
