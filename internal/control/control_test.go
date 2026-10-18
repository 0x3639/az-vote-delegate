package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/0x3639/az-vote-delegate/internal/config"
	"github.com/0x3639/az-vote-delegate/internal/delegate"
	"github.com/0x3639/znn-sdk-go/api/embedded"
)

var id = strings.Repeat("a", 64)

type fakeNode struct {
	owner   *delegate.Vote
	block   bool
	started chan struct{}
	sends   int
}

func (f *fakeNode) Check(context.Context) error { return nil }
func (f *fakeNode) Close()                      {}
func (f *fakeNode) Pillars(context.Context) (map[string]*embedded.PillarInfo, error) {
	return map[string]*embedded.PillarInfo{"A": {Name: "A"}, "B": {Name: "B"}, "C": {Name: "C"}}, nil
}
func (f *fakeNode) ListVotables(context.Context) ([]delegate.Votable, error) {
	return []delegate.Votable{{ID: id, Kind: "phase", Name: "Delivery payment"}}, nil
}
func (f *fakeNode) PillarVotes(_ context.Context, name string, _ []string) (delegate.Votes, error) {
	yes := delegate.Yes
	if name == "B" {
		return delegate.Votes{id: &yes}, nil
	}
	return delegate.Votes{id: f.owner}, nil
}
func (f *fakeNode) Executed(context.Context, string) (bool, error) { return true, nil }
func (f *fakeNode) Send(ctx context.Context, _ config.Delegation, _ string, v delegate.Vote, guard func() error, journal func(string) error) error {
	if f.block {
		close(f.started)
		<-ctx.Done()
		return ctx.Err()
	}
	if err := guard(); err != nil {
		return err
	}
	if err := journal(strings.Repeat("b", 64)); err != nil {
		return err
	}
	f.owner = &v
	f.sends++
	return nil
}
func manager(t *testing.T) (*Manager, *fakeNode) {
	t.Helper()
	dir := t.TempDir()
	c := config.Config{Live: true, ChainID: 1, StatePath: filepath.Join(dir, "state.json"), PollIntervalSeconds: 30, PowTimeoutSeconds: 30, Delegations: []config.Delegation{{PillarName: "A", Follow: "B"}}}
	f := &fakeNode{}
	m := &Manager{cfg: c, path: filepath.Join(dir, "config.json"), log: slog.New(slog.NewTextHandler(io.Discard, nil)), wake: make(chan struct{}, 1), fatal: make(chan error, 1)}
	m.factory = func(config.Config) (backend, error) { return f, nil }
	m.engine = &delegate.Engine{State: delegate.NewState(1), PollInterval: 30 * time.Second, PowTimeout: time.Minute, Save: func(s *delegate.State) error { return delegate.SaveState(c.StatePath, s) }}
	m.snapshot(nil, nil)
	return m, f
}
func command(t *testing.T, m *Manager, req Request) Response {
	t.Helper()
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, httptest.NewRequest("POST", "/command", bytes.NewReader(b)))
	var response Response
	if err = json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(rec.Body.String(), err)
	}
	return response
}
func TestUIInterruptsCopyAndPersistsManualIntent(t *testing.T) {
	m, f := manager(t)
	f.block = true
	f.started = make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- m.poll(context.Background()) }()
	select {
	case <-f.started:
	case <-time.After(3 * time.Second):
		t.Fatal("copy did not start")
	}
	// The preview remains readable while the signing worker is busy.
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, httptest.NewRequest("GET", "/status", nil))
	if !strings.Contains(rec.Body.String(), "Delivery payment") {
		t.Fatal("UI has no actionable preview during PoW")
	}
	response := command(t, m, Request{Action: "vote", Pillar: "A", ID: id, Vote: "no"})
	if response.Error != "" {
		t.Fatal(response.Error)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	state, err := delegate.LoadState(m.cfg.StatePath, 1)
	if err != nil {
		t.Fatal(err)
	}
	if state.Entries("A")[id].ManualVote == nil || *state.Entries("A")[id].ManualVote != delegate.No {
		t.Fatal("manual choice acknowledged without durable intent")
	}
	f.block = false
	if err = m.poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if f.sends != 1 || *f.owner != delegate.No {
		t.Fatal("copy won over UI manual vote")
	}
}
func TestPauseWorksWhenNodeUnavailable(t *testing.T) {
	m, _ := manager(t)
	m.factory = func(config.Config) (backend, error) { return nil, errors.New("offline") }
	response := command(t, m, Request{Action: "mode", Live: false})
	if response.Error != "" || !m.engine.DryRun || m.cfg.Live {
		t.Fatal("pause failed", response)
	}
}
func TestFailedManualSaveIsFatal(t *testing.T) {
	m, f := manager(t)
	m.engine.Save = func(*delegate.State) error { return errors.New("disk full") }
	response := command(t, m, Request{Action: "vote", Pillar: "A", ID: id, Vote: "abstain"})
	if response.Error == "" || f.sends != 0 {
		t.Fatal("manual failure acknowledged as success")
	}
	select {
	case err := <-m.fatal:
		if !errors.Is(err, delegate.ErrPersistence) {
			t.Fatal(err)
		}
	default:
		t.Fatal("state persistence error was not fatal")
	}
}
func TestFollowChangesPreserveManualOverride(t *testing.T) {
	m, _ := manager(t)
	vote := delegate.Yes
	m.engine.State.Entries("A")[id] = &delegate.Entry{ManualVote: &vote, OwnerOverride: true}
	response := command(t, m, Request{Action: "follow", Pillar: "A", Follow: "C"})
	if response.Error != "" || m.cfg.Delegations[0].Follow != "C" || m.engine.State.Entries("A")[id].ManualVote == nil {
		t.Fatal(response)
	}
}
