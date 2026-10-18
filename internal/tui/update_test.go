package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/0x3639/az-vote-delegate/internal/update"
)

type fakeReleases struct {
	checks, installs int
	err              error
}

func (f *fakeReleases) Latest(context.Context) (update.Release, error) {
	f.checks++
	return update.Release{Tag: "v0.2.6"}, f.err
}
func (f *fakeReleases) Install(context.Context, update.Release, string) error {
	f.installs++
	return nil
}
func TestUpdateNoticeAndCheckCaching(t *testing.T) {
	client := &fakeReleases{}
	u := &updater{current: "v0.2.5", client: client}
	u.check(context.Background(), false)
	if !u.available() || !strings.Contains(u.notice(), "Update available: v0.2.6") {
		t.Fatal(u.notice())
	}
	u.check(context.Background(), false)
	if client.checks != 1 {
		t.Fatal("menu return repeated the release request")
	}
	u.checked = time.Now().Add(-16 * time.Minute)
	u.check(context.Background(), false)
	if client.checks != 2 {
		t.Fatal("expired release check not refreshed")
	}
	client.err = errors.New("offline")
	u.check(context.Background(), true)
	if u.available() || !strings.Contains(u.notice(), "unavailable") {
		t.Fatal(u.notice())
	}
	if client.installs != 0 {
		t.Fatal("checking installed an update")
	}
}
func TestUpgradeBackDoesNotDownloadOrInstall(t *testing.T) {
	for _, keys := range []string{"\x1b", "\r"} {
		t.Run(keys, func(t *testing.T) {
			scriptedForms(t, keys)
			client := &fakeReleases{}
			u := &updater{current: "v0.2.5", client: client}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := u.upgrade(ctx, "/config.json"); !errors.Is(err, errBack) {
				t.Fatal(err)
			}
			if client.installs != 0 {
				t.Fatal("Back installed an update")
			}
		})
	}
}
