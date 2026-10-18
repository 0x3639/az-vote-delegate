// Package update checks and installs stable releases from the project's GitHub repository.
package update

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const Repository = "0x3639/az-vote-delegate"
const maxArchive = 64 << 20

var ErrCurrent = errors.New("the installed binary is already at this release or newer; reopen the menu")

type Release struct{ Tag string }

func (r Release) URL() string { return "https://github.com/" + Repository + "/releases/tag/" + r.Tag }

type Client struct {
	http           *http.Client
	api, downloads string
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 2 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" {
			return errors.New("release download redirected away from HTTPS")
		}
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		return nil
	}}, api: "https://api.github.com/repos/" + Repository + "/releases/latest", downloads: "https://github.com/" + Repository + "/releases/download/"}
}
func parseVersion(s string) ([3]uint64, error) {
	var v [3]uint64
	if !strings.HasPrefix(s, "v") || len(s) > 64 {
		return v, errors.New("expected a stable version such as v0.2.6")
	}
	parts := strings.Split(s[1:], ".")
	if len(parts) != 3 {
		return v, errors.New("invalid stable release version")
	}
	for i, p := range parts {
		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil || strconv.FormatUint(n, 10) != p {
			return v, errors.New("invalid stable release version")
		}
		v[i] = n
	}
	return v, nil
}
func Newer(latest, current string) (bool, error) {
	a, err := parseVersion(latest)
	if err != nil {
		return false, err
	}
	b, err := parseVersion(current)
	if err != nil {
		return false, err
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i], nil
		}
	}
	return false, nil
}
func (c *Client) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "az-vote-delegate-updater")
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		_ = res.Body.Close()
		return nil, fmt.Errorf("release server returned HTTP %d", res.StatusCode)
	}
	return res, nil
}
func (c *Client) Latest(ctx context.Context) (Release, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	res, err := c.get(ctx, c.api)
	if err != nil {
		return Release{}, err
	}
	defer func() { _ = res.Body.Close() }()
	var data struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&data); err != nil {
		return Release{}, err
	}
	if data.Draft || data.Prerelease {
		return Release{}, errors.New("latest release is not a published stable release")
	}
	if _, err = parseVersion(data.Tag); err != nil {
		return Release{}, err
	}
	return Release{Tag: data.Tag}, nil
}
func checksum(data []byte, asset string) (string, error) {
	found := ""
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[1] != asset {
			continue
		}
		bytes, err := hex.DecodeString(fields[0])
		if err != nil || len(bytes) != sha256.Size || found != "" {
			return "", errors.New("invalid or duplicate release checksum")
		}
		found = strings.ToLower(fields[0])
	}
	if found == "" {
		return "", errors.New("release checksum is missing")
	}
	return found, nil
}
func versionOf(ctx context.Context, path string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "version").Output()
	if err != nil {
		return "", fmt.Errorf("could not verify executable version: %w", err)
	}
	value := strings.TrimSpace(string(out))
	if !strings.HasPrefix(value, "az-vote-delegate ") {
		return "", errors.New("unexpected executable version output")
	}
	value = strings.TrimPrefix(value, "az-vote-delegate ")
	_, err = parseVersion(value)
	return value, err
}
func Executable() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(p)
}
func (c *Client) Install(ctx context.Context, r Release, path string) error {
	if runtime.GOOS != "linux" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		return errors.New("in-app upgrades support Linux amd64 and arm64")
	}
	return c.install(ctx, r, path, runtime.GOARCH)
}
func (c *Client) install(ctx context.Context, r Release, path, arch string) error {
	if _, err := parseVersion(r.Tag); err != nil {
		return err
	}
	if arch != "amd64" && arch != "arm64" {
		return errors.New("unsupported architecture")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("upgrade target must be a regular executable")
	}
	// Serialize concurrent menus. Keep the lock inode stable across upgrades.
	fd, err := unix.Open(path+".update.lock", unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return fmt.Errorf("cannot write installation directory; run the menu as its owner: %w", err)
	}
	defer func() { _ = unix.Close(fd) }()
	if err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return errors.New("another upgrade is already running")
	}
	current, err := versionOf(ctx, path)
	if err != nil {
		return err
	}
	newer, err := Newer(r.Tag, current)
	if err != nil {
		return err
	}
	if !newer {
		return ErrCurrent
	}
	stage, err := os.CreateTemp(filepath.Dir(path), ".az-upgrade-*")
	if err != nil {
		return err
	}
	defer func() { _ = stage.Close(); _ = os.Remove(stage.Name()) }()
	archive, err := os.CreateTemp("", "az-download-*")
	if err != nil {
		return err
	}
	defer func() { _ = archive.Close(); _ = os.Remove(archive.Name()) }()
	asset := "az-vote-delegate_linux_" + arch + ".tar.gz"
	base := c.downloads + r.Tag + "/"
	res, err := c.get(ctx, base+"checksums.txt")
	if err != nil {
		return err
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	_ = res.Body.Close()
	if err != nil {
		return err
	}
	if len(data) > 1<<20 {
		return errors.New("checksum file is too large")
	}
	expected, err := checksum(data, asset)
	if err != nil {
		return err
	}
	res, err = c.get(ctx, base+asset)
	if err != nil {
		return err
	}
	digest := sha256.New()
	n, err := io.Copy(io.MultiWriter(archive, digest), io.LimitReader(res.Body, maxArchive+1))
	_ = res.Body.Close()
	if err != nil {
		return err
	}
	if n > maxArchive {
		return errors.New("release archive is too large")
	}
	if hex.EncodeToString(digest.Sum(nil)) != expected {
		return errors.New("checksum verification failed; existing executable preserved")
	}
	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err = extract(archive, stage); err != nil {
		return err
	}
	if err = stage.Chmod(0755); err != nil {
		return err
	}
	if err = stage.Sync(); err != nil {
		return err
	}
	if err = stage.Close(); err != nil {
		return err
	}
	stagedVersion, err := versionOf(ctx, stage.Name())
	if err != nil {
		return err
	}
	if stagedVersion != r.Tag {
		return errors.New("downloaded executable version does not match release")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	// Same-directory rename leaves either the entire old binary or the new one.
	if err = os.Rename(stage.Name(), path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("binary installed but cannot sync directory: %w", err)
	}
	defer func() { _ = dir.Close() }()
	if err = dir.Sync(); err != nil {
		return fmt.Errorf("binary installed but directory sync failed: %w", err)
	}
	return nil
}
func extract(src io.Reader, dst io.Writer) error {
	gz, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(io.LimitReader(gz, 128<<20))
	found := false
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if h.Name != "az-vote-delegate" {
			continue
		}
		if found || h.Typeflag != tar.TypeReg || h.Size <= 0 || h.Size > maxArchive {
			return errors.New("invalid executable in release archive")
		}
		if _, err = io.CopyN(dst, tr, h.Size); err != nil {
			return err
		}
		found = true
	}
	if !found {
		return errors.New("release archive has no executable")
	}
	return nil
}
