package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/0x3639/az-vote-delegate/internal/config"
	"github.com/0x3639/az-vote-delegate/internal/control"
	"github.com/0x3639/az-vote-delegate/internal/delegate"
	"github.com/0x3639/az-vote-delegate/internal/node"
	"github.com/0x3639/az-vote-delegate/internal/setup"
	"github.com/0x3639/az-vote-delegate/internal/tui"
	"github.com/charmbracelet/huh"
)

var version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go func() { <-ctx.Done(); cancel() }() // Restore normal signals after the first interrupt.
	if err := execute(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil && ctx.Err() == nil && !errors.Is(err, huh.ErrUserAborted) {
		fmt.Fprintln(os.Stderr, "az-vote-delegate:", err)
		os.Exit(1)
	}
}
func execute(ctx context.Context, args []string, out, stderr io.Writer) error {
	if len(args) == 0 {
		args = []string{"ui"}
	}
	if args[0] == "version" {
		if len(args) != 1 {
			return errors.New("version takes no arguments")
		}
		_, err := fmt.Fprintln(out, "az-vote-delegate", version)
		return err
	}
	if args[0] == "--help" || args[0] == "help" {
		_, err := fmt.Fprintln(out, "az-vote-delegate [ui|setup|run|status|check|service-install|version] [--config path]\nRun without arguments for the interactive menu.")
		return err
	}
	command := args[0]
	if command != "run" && command != "status" && command != "check" && command != "ui" && command != "setup" && command != "service-install" {
		return fmt.Errorf("unknown command %q", command)
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("config", config.DefaultPath(), "configuration file")
	debug := flags.Bool("debug", false, "include unchanged decisions in logs")
	dryRun := true
	reset := false
	if command == "run" {
		flags.BoolVar(&dryRun, "dry-run", true, "log decisions without publishing (use --dry-run=false for live votes)")
		flags.BoolVar(&reset, "reset-state", false, "discard saved vote attribution and pending sends before running")
	}
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if command == "ui" {
		return tui.Run(ctx, *path, version)
	}
	if command == "setup" {
		return tui.Wizard(ctx, *path)
	}
	if command == "service-install" {
		if _, err := config.Load(*path); err != nil {
			return err
		}
		return setup.InstallService(*path)
	}
	if reset && dryRun {
		return errors.New("--reset-state requires live mode (--dry-run=false)")
	}
	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level}))
	c, err := config.Load(*path)
	if err != nil {
		return err
	}
	explicitDry := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "dry-run" {
			explicitDry = true
		}
	})
	if !explicitDry {
		dryRun = !c.Live
	}
	if command == "status" {
		if res, err := control.Call(ctx, c, nil); err == nil && res.Snapshot != nil {
			if res.Snapshot.Error != "" {
				return errors.New(res.Snapshot.Error)
			}
			return printStatus(out, res.Snapshot.Rows)
		}
	}
	keys, err := node.LoadKeys(c.Delegations)
	if err != nil {
		return err
	}
	defer keys.Destroy()
	if command == "check" {
		n, err := node.Open(c, keys, log)
		if err != nil {
			return err
		}
		defer n.Close()
		err = n.Check(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "check passed: chainId=%d, %d delegation(s), producer keys match active pillars\n", c.ChainID, len(c.Delegations))
		return err
	}
	if command == "run" {
		return control.Run(ctx, *path, c, keys, dryRun, log, reset)
	}
	state, err := delegate.LoadState(c.StatePath, c.ChainID)
	if err != nil {
		return err
	}
	n, err := node.Open(c, keys, log)
	if err != nil {
		return err
	}
	defer n.Close()
	if err = n.Check(ctx); err != nil {
		return err
	}
	engine := &delegate.Engine{Backend: n, State: state, Log: log, CopyDelay: c.CopyDelay(), PollInterval: c.PollInterval(), PowTimeout: c.PowTimeout(), DryRun: true}
	rows, err := engine.Reconcile(ctx, c.Delegations)
	if err != nil {
		return err
	}
	return printStatus(out, rows)
}

func printStatus(out io.Writer, rows []delegate.Row) error {
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "PILLAR\tFOLLOW\tKIND\tVOTABLE\tDELEGATE\tOWNER\tACTION\tCOPY AFTER (UTC)"); err != nil {
		return err
	}
	for _, r := range rows {
		copyAfter := "-"
		if !r.CopyAfter.IsZero() {
			copyAfter = r.CopyAfter.UTC().Format(time.RFC3339)
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", r.Pillar, r.Follow, r.Votable.Kind, r.Votable.ID, delegate.VoteText(r.Delegate), delegate.VoteText(r.Owner), r.Action, copyAfter); err != nil {
			return err
		}
	}
	return w.Flush()
}
