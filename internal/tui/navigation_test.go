package tui

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0x3639/az-vote-delegate/internal/config"
	"github.com/0x3639/az-vote-delegate/internal/control"
	"github.com/0x3639/az-vote-delegate/internal/delegate"
	"github.com/charmbracelet/huh"
)

func keyboardForm(ctx context.Context, keys string, fields ...huh.Field) error {
	f, back := navigationForm(fields...)
	f.WithAccessible(false).WithInput(strings.NewReader(keys)).WithOutput(io.Discard)
	err := f.RunWithContext(ctx)
	return formResult(err, *back)
}

func TestEscapeAndInterrupt(t *testing.T) {
	for _, tc := range []struct {
		name, keys string
		filtering  bool
		want       error
	}{
		{"escape", "\x1b", false, errBack},
		{"escape while filtering", "\x1b", true, errBack},
		{"interrupt", "\x03", false, huh.ErrUserAborted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			value := "one"
			err := keyboardForm(ctx, tc.keys, huh.NewSelect[string]().Options(huh.NewOption("One", "one")).Filtering(tc.filtering).Value(&value))
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func scriptedForms(t *testing.T, scripts ...string) {
	t.Helper()
	original := runForm
	count := 0
	runForm = func(ctx context.Context, fields ...huh.Field) error {
		if count >= len(scripts) {
			t.Fatal("unexpected additional screen")
		}
		keys := scripts[count]
		count++
		return keyboardForm(ctx, keys, fields...)
	}
	t.Cleanup(func() {
		runForm = original
		if count != len(scripts) {
			t.Errorf("showed %d screens, want %d", count, len(scripts))
		}
	})
}

func TestModeBackDoesNotSwitch(t *testing.T) {
	scriptedForms(t, "\x1b[B\r", "\x1b", "\x1b") // Live, Back at confirmation, Back at picker.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// No service exists: any attempted mutation would fail instead of returning Back.
	err := changeMode(ctx, config.Config{StatePath: filepath.Join(t.TempDir(), "state.json")}, false)
	if !errors.Is(err, errBack) {
		t.Fatalf("got %v, want Back", err)
	}
}

func TestRequestBackNavigationDoesNotVote(t *testing.T) {
	for _, tc := range []struct {
		name    string
		scripts []string
	}{
		{"escape from details", []string{"\r", "\x1b", "\x1b"}},
		{"escape confirmation then visible Back", []string{"\r", "\r", "\x1b", "\x1b[B\x1b[B\x1b[B\r", "\x1b[B\r"}},
		{"Back confirmation then escape", []string{"\r", "\r", "\r", "\x1b", "\x1b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scriptedForms(t, tc.scripts...)
			// Keep Unix socket paths below the macOS length limit.
			dir, err := os.MkdirTemp("", "az-nav-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.RemoveAll(dir); err != nil {
					t.Error(err)
				}
			})
			c := config.Config{StatePath: filepath.Join(dir, "state.json")}
			listener, err := net.Listen("unix", control.SocketPath(c))
			if err != nil {
				t.Fatal(err)
			}
			var reads, mutations atomic.Int32
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req control.Request
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Error(err)
					return
				}
				if req.Action == "requests" {
					reads.Add(1)
				} else {
					mutations.Add(1)
				}
				_ = json.NewEncoder(w).Encode(control.Response{Snapshot: &control.Snapshot{Live: true, Rows: []delegate.Row{{Pillar: "own", Votable: delegate.Votable{ID: "request", Kind: "phase", Name: "Phase 1"}}}}})
			}))
			_ = server.Listener.Close()
			server.Listener = listener
			server.Start()
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = votables(ctx, c, true)
			if !errors.Is(err, errBack) {
				t.Fatalf("got %v, want Back", err)
			}
			if reads.Load() != 2 || mutations.Load() != 0 {
				t.Fatalf("reads=%d mutations=%d; want two list reads and no mutation", reads.Load(), mutations.Load())
			}
		})
	}
}
