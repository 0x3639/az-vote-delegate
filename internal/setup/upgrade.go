package setup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const serviceName = "az-vote-delegate.service"

type systemCommand func(context.Context, ...string) ([]byte, error)

type UpgradeService struct {
	Restart        bool
	Description    string
	binary, config string
	user           bool
	run            systemCommand
	read           func(string) ([]byte, error)
}

func runSystemctl(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "systemctl", args...).CombinedOutput()
}
func PlanUpgrade(ctx context.Context, binary, path string) UpgradeService {
	if runtime.GOOS != "linux" {
		return UpgradeService{Description: "In-app upgrades require Linux."}
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return UpgradeService{Description: "Restart the service manually after upgrading; configuration path could not be resolved."}
	}
	return inspectUpgrade(ctx, binary, path, os.Geteuid() != 0, runSystemctl, os.ReadFile)
}
func inspectUpgrade(ctx context.Context, binary, path string, user bool, run systemCommand, read func(string) ([]byte, error)) UpgradeService {
	p := UpgradeService{binary: binary, config: path, user: user, run: run, read: read, Description: "Restart your existing foreground process or custom service manually after upgrading. No service will be created."}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	args := []string{"show", serviceName, "--property=LoadState,ActiveState,FragmentPath,DropInPaths,NeedDaemonReload"}
	if user {
		args = append([]string{"--user"}, args...)
	}
	out, err := run(ctx, args...)
	if err != nil {
		return p
	}
	props := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if ok {
			props[k] = v
		}
	}
	if props["LoadState"] != "loaded" {
		return p
	}
	if props["ActiveState"] == "inactive" || props["ActiveState"] == "failed" {
		p.Description = "The existing systemd service is stopped and will stay stopped. No service will be created."
		return p
	}
	if props["ActiveState"] != "active" || props["DropInPaths"] != "" || props["FragmentPath"] == "" || props["NeedDaemonReload"] != "no" {
		return p
	}
	expected, err := Unit(binary, path, user)
	if err != nil {
		return p
	}
	actual, err := read(props["FragmentPath"])
	if err != nil || string(actual) != expected {
		return p
	}
	p.Restart = true
	p.Description = "Restart the existing az-vote-delegate service after installation. Your voting mode and manual overrides stay in place."
	return p
}
func (p UpgradeService) RestartExisting(ctx context.Context) error {
	if !p.Restart {
		return nil
	}
	// Recheck the unit before touching it. Never install or enable a service here.
	current := inspectUpgrade(ctx, p.binary, p.config, p.user, p.run, p.read)
	if !current.Restart {
		return fmt.Errorf("binary upgraded, but service configuration or status changed; restart your existing service manually")
	}
	ctx, cancel := context.WithTimeout(ctx, 55*time.Second)
	defer cancel()
	args := []string{"try-restart", serviceName}
	if p.user {
		args = append([]string{"--user"}, args...)
	}
	out, err := p.run(ctx, args...)
	if err != nil {
		return fmt.Errorf("binary upgraded, but service restart failed: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	args = []string{"is-active", "--quiet", serviceName}
	if p.user {
		args = append([]string{"--user"}, args...)
	}
	if _, err = p.run(ctx, args...); err != nil {
		return fmt.Errorf("binary upgraded, but service is not active; inspect its logs and restart it: %w", err)
	}
	return nil
}
