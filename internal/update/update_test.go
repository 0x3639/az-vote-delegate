package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestNewer(t *testing.T) {
	for _, tc := range []struct {
		latest, current string
		want, bad       bool
	}{
		{"v0.2.10", "v0.2.9", true, false}, {"v1.0.0", "v0.99.99", true, false},
		{"v0.2.6", "v0.2.6", false, false}, {"v0.2.5", "v0.2.6", false, false},
		{"v0.2.6", "dev", false, true}, {"v0.2.6-rc1", "v0.2.5", false, true},
		{"v01.2.3", "v1.2.2", false, true}, {"../../evil", "v0.2.5", false, true},
	} {
		got, err := Newer(tc.latest, tc.current)
		if got != tc.want || (err != nil) != tc.bad {
			t.Errorf("%s vs %s: %v, %v", tc.latest, tc.current, got, err)
		}
	}
}
func TestLatest(t *testing.T) {
	for _, tc := range []struct {
		body   string
		status int
		bad    bool
	}{
		{`{"tag_name":"v0.2.6"}`, 200, false},
		{`{"tag_name":"v0.2.6","draft":true}`, 200, true},
		{`{"tag_name":"v0.2.6","prerelease":true}`, 200, true},
		{`{"tag_name":"v0.2.6/other"}`, 200, true},
		{`{"tag_name":`, 200, true}, {`rate limited`, 403, true},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "" {
				t.Error("release check must not send credentials")
			}
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		}))
		c := &Client{http: server.Client(), api: server.URL}
		got, err := c.Latest(context.Background())
		server.Close()
		if (err != nil) != tc.bad {
			t.Fatalf("%s: %v", tc.body, err)
		}
		if !tc.bad && got.Tag != "v0.2.6" {
			t.Fatal(got)
		}
	}
}
func executable(version string) []byte {
	return []byte("#!/bin/sh\necho 'az-vote-delegate " + version + "'\n")
}
func archive(t *testing.T, name string, kind byte, body []byte, duplicate bool) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	size := int64(len(body))
	if kind != tar.TypeReg {
		size = 0
	}
	header := &tar.Header{Name: name, Mode: 0755, Typeflag: kind, Size: size, Linkname: "/elsewhere"}
	count := 1
	if duplicate {
		count = 2
	}
	for range count {
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if size > 0 {
			if _, err := tw.Write(body); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func releaseServer(t *testing.T, data []byte, corrupt, missing bool) *Client {
	t.Helper()
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if corrupt {
		hash = strings.Repeat("0", 64)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if missing {
			http.NotFound(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "checksums.txt") {
			_, _ = fmt.Fprintf(w, "%s  az-vote-delegate_linux_amd64.tar.gz\n%s  az-vote-delegate_linux_arm64.tar.gz\n", hash, hash)
			return
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	return &Client{http: server.Client(), downloads: server.URL + "/"}
}
func installed(t *testing.T) (string, []byte) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "az-vote-delegate")
	old := executable("v0.2.5")
	if err := os.WriteFile(path, old, 0755); err != nil {
		t.Fatal(err)
	}
	return path, old
}
func TestInstallReplacesOnlyExecutable(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			binary, _ := installed(t)
			state := filepath.Join(filepath.Dir(binary), "state.json")
			saved := []byte(`{"manualVote":1,"manualConfirmed":true}`)
			if err := os.WriteFile(state, saved, 0600); err != nil {
				t.Fatal(err)
			}
			data := archive(t, "az-vote-delegate", tar.TypeReg, executable("v0.2.6"), false)
			c := releaseServer(t, data, false, false)
			if err := c.install(context.Background(), Release{Tag: "v0.2.6"}, binary, arch); err != nil {
				t.Fatal(err)
			}
			got, err := versionOf(context.Background(), binary)
			if err != nil || got != "v0.2.6" {
				t.Fatal(got, err)
			}
			preserved, err := os.ReadFile(state)
			if err != nil || !bytes.Equal(preserved, saved) {
				t.Fatal("state changed", err)
			}
			if err := c.install(context.Background(), Release{Tag: "v0.2.6"}, binary, arch); err != ErrCurrent {
				t.Fatal("repeat install", err)
			}
			if err := c.install(context.Background(), Release{Tag: "v0.2.5"}, binary, arch); err != ErrCurrent {
				t.Fatal("downgrade allowed", err)
			}
			leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(binary), ".az-upgrade-*"))
			if len(leftovers) != 0 {
				t.Fatal("temporary binaries retained")
			}
		})
	}
}
func TestFailedUpgradePreservesExecutable(t *testing.T) {
	valid := archive(t, "az-vote-delegate", tar.TypeReg, executable("v0.2.6"), false)
	for _, tc := range []struct {
		name             string
		data             []byte
		corrupt, missing bool
	}{
		{"checksum", valid, true, false}, {"missing release", valid, false, true},
		{"invalid gzip", []byte("invalid"), false, false},
		{"symlink", archive(t, "az-vote-delegate", tar.TypeSymlink, nil, false), false, false},
		{"wrong version", archive(t, "az-vote-delegate", tar.TypeReg, executable("v0.2.4"), false), false, false},
		{"wrong path", archive(t, "../az-vote-delegate", tar.TypeReg, executable("v0.2.6"), false), false, false},
		{"duplicate executable", archive(t, "az-vote-delegate", tar.TypeReg, executable("v0.2.6"), true), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			binary, old := installed(t)
			c := releaseServer(t, tc.data, tc.corrupt, tc.missing)
			if err := c.install(context.Background(), Release{Tag: "v0.2.6"}, binary, "amd64"); err == nil {
				t.Fatal("invalid upgrade accepted")
			}
			got, err := os.ReadFile(binary)
			if err != nil || !bytes.Equal(got, old) {
				t.Fatal("existing binary changed", err)
			}
		})
	}
}
func TestUpgradeLockAndCancellation(t *testing.T) {
	binary, old := installed(t)
	c := releaseServer(t, archive(t, "az-vote-delegate", tar.TypeReg, executable("v0.2.6"), false), false, false)
	lock, err := os.OpenFile(binary+".update.lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Close() }()
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	if err = c.install(context.Background(), Release{Tag: "v0.2.6"}, binary, "amd64"); err == nil || !strings.Contains(err.Error(), "another upgrade") {
		t.Fatal(err)
	}
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_UN); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err = c.install(ctx, Release{Tag: "v0.2.6"}, binary, "amd64"); err == nil {
		t.Fatal("canceled upgrade succeeded")
	}
	got, err := os.ReadFile(binary)
	if err != nil || !bytes.Equal(got, old) {
		t.Fatal("existing binary changed", err)
	}
}

func TestChecksumRequiresSingleMatchingEntry(t *testing.T) {
	asset := "az-vote-delegate_linux_amd64.tar.gz"
	hash := strings.Repeat("a", 64)
	for _, body := range []string{"", hash + "  different.tar.gz", "invalid  " + asset, hash + "  " + asset + "\n" + hash + "  " + asset} {
		if _, err := checksum([]byte(body), asset); err == nil {
			t.Fatalf("accepted %q", body)
		}
	}
	if got, err := checksum([]byte(hash+"  "+asset), asset); err != nil || got != hash {
		t.Fatal(got, err)
	}
}
