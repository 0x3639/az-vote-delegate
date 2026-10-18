package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/0x3639/az-vote-delegate/internal/config"
)

type Producer struct {
	KeyFilePath string
	Index       int
	Password    string
}

func ReadProducer(path string) (Producer, error) {
	f, err := os.Open(path)
	if err != nil {
		return Producer{}, err
	}
	defer func() { _ = f.Close() }()
	var doc struct{ Producer *Producer }
	if err = json.NewDecoder(io.LimitReader(f, 1<<20)).Decode(&doc); err != nil {
		return Producer{}, errors.New("cannot parse znnd configuration")
	}
	if doc.Producer == nil || doc.Producer.KeyFilePath == "" {
		return Producer{}, errors.New("znnd config has no producer key")
	}
	p := *doc.Producer
	if !filepath.IsAbs(p.KeyFilePath) {
		p.KeyFilePath = filepath.Join(filepath.Dir(path), "wallet", p.KeyFilePath)
	}
	return p, nil
}
func Unit(binary, path string, user bool) (string, error) {
	for _, s := range []string{binary, path} {
		if !filepath.IsAbs(s) || strings.ContainsAny(s, "\r\n\x00") {
			return "", errors.New("service paths must be absolute without newlines")
		}
	}
	quote := func(s string) string { return strconv.Quote(strings.ReplaceAll(s, "%", "%%")) }
	target := "multi-user.target"
	if user {
		target = "default.target"
	}
	return fmt.Sprintf(`[Unit]
Description=Accelerator-Z vote manager
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s run --config %s
Restart=on-failure
RestartSec=30
TimeoutStopSec=45
UMask=0077
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=%s
`, quote(binary), quote(path), target), nil
}
func InstallService(path string) error {
	if runtime.GOOS != "linux" {
		return errors.New("background service installation requires Linux with systemd; use run in a second terminal here")
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	binary, err = filepath.EvalSymlinks(binary)
	if err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	user := os.Geteuid() != 0
	dir := "/etc/systemd/system"
	args := []string{}
	if user {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(home, ".config", "systemd", "user")
		args = append(args, "--user")
	}
	// Check systemd availability before leaving a unit file behind.
	probe := exec.Command("systemctl", append(append([]string{}, args...), "show-environment")...)
	if out, err := probe.CombinedOutput(); err != nil {
		return fmt.Errorf("systemd unavailable: %s", strings.TrimSpace(string(out)))
	}
	unit, err := Unit(binary, path, user)
	if err != nil {
		return err
	}
	if err = config.WritePrivate(filepath.Join(dir, "az-vote-delegate.service"), []byte(unit)); err != nil {
		return err
	}
	for _, command := range [][]string{{"daemon-reload"}, {"enable", "--now", "az-vote-delegate.service"}} {
		out, err := exec.Command("systemctl", append(append([]string{}, args...), command...)...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl: %s", strings.TrimSpace(string(out)))
		}
	}
	return nil
}
