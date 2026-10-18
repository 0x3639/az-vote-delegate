package setup

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestUpgradeRestartsOnlyMatchingActiveService(t *testing.T) {
	for _, tc := range []struct {
		name, state, load, dropins string
		custom, user, want, reload bool
	}{
		{name: "system service", state: "active", load: "loaded", want: true},
		{name: "user service", state: "active", load: "loaded", user: true, want: true},
		{name: "stopped", state: "inactive", load: "loaded"},
		{name: "failed", state: "failed", load: "loaded"},
		{name: "missing", state: "inactive", load: "not-found"},
		{name: "different config", state: "active", load: "loaded", custom: true},
		{name: "drop-in", state: "active", load: "loaded", dropins: "/override.conf"},
		{name: "pending daemon reload", state: "active", load: "loaded", reload: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, err := Unit("/usr/local/bin/az-vote-delegate", "/config.json", tc.user)
			if err != nil {
				t.Fatal(err)
			}
			if tc.custom {
				unit = strings.ReplaceAll(unit, "/config.json", "/other.json")
			}
			restarted := 0
			reload := "no"
			if tc.reload {
				reload = "yes"
			}
			run := func(_ context.Context, args ...string) ([]byte, error) {
				if tc.user {
					if args[0] != "--user" {
						t.Fatal("wrong service scope", args)
					}
					args = args[1:]
				}
				switch args[0] {
				case "show":
					return []byte("LoadState=" + tc.load + "\nActiveState=" + tc.state + "\nFragmentPath=/unit\nNeedDaemonReload=" + reload + "\nDropInPaths=" + tc.dropins + "\n"), nil
				case "try-restart":
					if len(args) != 2 || args[1] != serviceName {
						t.Fatal(args)
					}
					restarted++
					return nil, nil
				case "is-active":
					return nil, nil
				default:
					t.Fatal("unexpected service mutation", args)
					return nil, nil
				}
			}
			plan := inspectUpgrade(context.Background(), "/usr/local/bin/az-vote-delegate", "/config.json", tc.user, run, func(string) ([]byte, error) { return []byte(unit), nil })
			if plan.Restart != tc.want {
				t.Fatal(plan)
			}
			if err = plan.RestartExisting(context.Background()); err != nil {
				t.Fatal(err)
			}
			want := 0
			if tc.want {
				want = 1
			}
			if restarted != want {
				t.Fatalf("restarts=%d, want %d", restarted, want)
			}
		})
	}
}
func TestUpgradeRechecksServiceAndReportsRestartFailure(t *testing.T) {
	for _, change := range []bool{false, true} {
		unit, _ := Unit("/binary", "/config", false)
		shows := 0
		run := func(_ context.Context, args ...string) ([]byte, error) {
			if args[0] == "show" {
				shows++
				return []byte("LoadState=loaded\nActiveState=active\nFragmentPath=/unit\nNeedDaemonReload=no\nDropInPaths=\n"), nil
			}
			if change {
				t.Fatal("changed service was restarted")
			}
			return []byte("restart failed"), errors.New("test failure")
		}
		read := func(string) ([]byte, error) {
			if change && shows > 1 {
				return []byte("different config"), nil
			}
			return []byte(unit), nil
		}
		plan := inspectUpgrade(context.Background(), "/binary", "/config", false, run, read)
		if !plan.Restart {
			t.Fatal("expected matching service")
		}
		if err := plan.RestartExisting(context.Background()); err == nil || !strings.Contains(err.Error(), "binary upgraded") {
			t.Fatal("partial failure not reported", err)
		}
	}
}
