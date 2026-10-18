package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/0x3639/az-vote-delegate/internal/setup"
	"github.com/0x3639/az-vote-delegate/internal/update"
	"github.com/charmbracelet/huh"
)

type releaseClient interface {
	Latest(context.Context) (update.Release, error)
	Install(context.Context, update.Release, string) error
}
type updater struct {
	current string
	client  releaseClient
	release update.Release
	err     error
	checked time.Time
}

func (u *updater) check(ctx context.Context, force bool) {
	if !force && !u.checked.IsZero() && time.Since(u.checked) < 15*time.Minute {
		return
	}
	u.checked = time.Now()
	u.release, u.err = u.client.Latest(ctx)
}
func (u *updater) available() bool {
	if u.err != nil {
		return false
	}
	newer, err := update.Newer(u.release.Tag, u.current)
	return err == nil && newer
}
func (u *updater) notice() string {
	if u.available() {
		return "Update available: " + u.release.Tag + " (installed " + Clean(u.current) + "). Choose Upgrade application."
	}
	if u.err != nil {
		return "Installed " + Clean(u.current) + " · update check unavailable; try Check for updates."
	}
	return "Installed " + Clean(u.current)
}
func (u *updater) option() huh.Option[string] {
	label := "Check for updates / upgrade"
	if u.available() {
		label = "Upgrade application to " + u.release.Tag
	}
	return huh.NewOption(label, "upgrade")
}
func (u *updater) upgrade(ctx context.Context, path string) error {
	fmt.Fprintln(os.Stderr, "Checking the latest stable release…")
	u.check(ctx, true)
	if u.err != nil {
		return fmt.Errorf("could not check for updates; voting continues unchanged: %w", u.err)
	}
	newer, err := update.Newer(u.release.Tag, u.current)
	if err != nil {
		return errors.New("this build has no stable release version; install a published release with the curl installer")
	}
	if !newer {
		return note(ctx, "Application is up to date", fmt.Sprintf("Installed: %s\nLatest stable release: %s\nNo newer stable release is available.", Clean(u.current), u.release.Tag))
	}
	binary, err := update.Executable()
	if err != nil {
		return fmt.Errorf("reopen the menu before upgrading: %w", err)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	service := setup.PlanUpgrade(ctx, binary, path)
	description := fmt.Sprintf("%s → %s\n%s\n\n%s\n\nConfiguration, passwords and saved manual overrides are preserved.\nRelease notes: %s", Clean(u.current), u.release.Tag, Clean(binary), service.Description, u.release.URL())
	yes := false
	if err = form(ctx, huh.NewConfirm().Title("Upgrade application?").Description(description).Affirmative("Upgrade").Negative("Back").Value(&yes)); err != nil {
		return err
	}
	if !yes {
		return errBack
	}
	fmt.Fprintln(os.Stderr, "Downloading and verifying the release. Voting continues during the download…")
	if err = u.client.Install(ctx, u.release, binary); err != nil {
		return err
	}
	message := "Installed " + u.release.Tag + ". Your settings and manual overrides are preserved."
	if service.Restart {
		fmt.Fprintln(os.Stderr, "Restarting the existing voting service…")
		if err = service.RestartExisting(ctx); err != nil {
			// The binary is already upgraded. Explain the partial outcome and reopen
			// the new UI rather than attempting to install it a second time.
			message += "\n\n" + Clean(err.Error())
		} else {
			message += "\nThe existing voting service was restarted."
		}
	} else {
		message += "\n\n" + service.Description
	}
	if err = form(ctx, huh.NewNote().Title("Application upgraded").Description(message).Next(true).NextLabel("Reopen menu")); err != nil {
		if errors.Is(err, errBack) {
			return huh.ErrUserAborted
		}
		return err
	}
	// Replace this UI process. Reopening never creates a daemon or changes mode.
	err = syscall.Exec(binary, []string{binary, "ui", "--config", path}, os.Environ())
	return fmt.Errorf("%s is installed, but the menu could not reopen; run %s ui --config %q: %w", u.release.Tag, binary, path, err)
}
