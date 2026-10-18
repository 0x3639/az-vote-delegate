package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/0x3639/az-vote-delegate/internal/config"
	"github.com/0x3639/az-vote-delegate/internal/control"
	"github.com/0x3639/az-vote-delegate/internal/delegate"
	"github.com/0x3639/az-vote-delegate/internal/node"
	"github.com/0x3639/az-vote-delegate/internal/setup"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// Clean strips terminal controls from names and descriptions read from chain.
func Clean(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == 0x202e || r == 0x202d || (r >= 0x2066 && r <= 0x2069) {
			return -1
		}
		return r
	}, s)
}
func form(ctx context.Context, fields ...huh.Field) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	theme := huh.ThemeBase()
	theme.Focused.Title = theme.Focused.Title.Foreground(lipgloss.Color("46"))
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(lipgloss.Color("46"))
	return huh.NewForm(huh.NewGroup(fields...)).WithTheme(theme).RunWithContext(ctx)
}
func confirm(ctx context.Context, title string) (bool, error) {
	yes := false
	err := form(ctx, huh.NewConfirm().Title(title).Value(&yes))
	return yes, err
}
func note(ctx context.Context, title, body string) error {
	return form(ctx, huh.NewNote().Title(title).Description(body).Next(true).NextLabel("Continue"))
}
func required(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("required")
	}
	return nil
}
func absolute(s string) error {
	if !filepath.IsAbs(s) {
		return errors.New("use an absolute path")
	}
	return nil
}
func Run(ctx context.Context, path string) error {
	if !term.IsTerminal(os.Stdin.Fd()) || !term.IsTerminal(os.Stderr.Fd()) {
		return errors.New("the interactive menu needs a terminal; use --help for commands")
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err = Wizard(ctx, path); err != nil {
			return err
		}
	}
	for ctx.Err() == nil {
		c, err := config.Load(path)
		if err != nil {
			return err
		}
		response, err := control.Call(ctx, c, nil)
		if err != nil {
			action := ""
			if err = form(ctx, huh.NewSelect[string]().Title("AZ vote manager — service is stopped").Description("Start the service to preview votes, select a delegate and vote manually.").Options(huh.NewOption("Install and start background service (systemd)", "service"), huh.NewOption("Show foreground command", "foreground"), huh.NewOption("Retry connection", "retry"), huh.NewOption("Exit", "exit")).Value(&action)); err != nil {
				return err
			}
			switch action {
			case "exit":
				return nil
			case "service":
				if err = setup.InstallService(path); err != nil {
					if err = note(ctx, "Could not start service", Clean(err.Error())); err != nil {
						return err
					}
				} else {
					time.Sleep(300 * time.Millisecond)
				}
			case "foreground":
				if err = note(ctx, "Run in a second terminal", fmt.Sprintf("az-vote-delegate run --config %q\n\nThen choose Retry connection here. For a persistent systemd user service, enable lingering with loginctl enable-linger USER.", path)); err != nil {
					return err
				}
			}
			continue
		}
		s := response.Snapshot
		if s == nil {
			return errors.New("empty service status")
		}
		mode := "DRY RUN — previews only"
		if s.Live {
			mode = "LIVE — votes can be published"
		}
		action := ""
		options := []huh.Option[string]{huh.NewOption("Project proposals", "projects"), huh.NewOption("Phase payouts / payment requests", "phases"), huh.NewOption("Choose pillar to copy", "follow"), huh.NewOption("Pending transactions & manual overrides", "activity"), huh.NewOption("Pause / enable live voting", "mode"), huh.NewOption("Refresh", "refresh"), huh.NewOption("Exit (service keeps running)", "exit")}
		description := fmt.Sprintf("%s  ·  Chain %d\n", mode, s.ChainID)
		for _, d := range s.Delegations {
			description += Clean(d.Pillar) + " follows " + Clean(d.Follow) + "\n"
		}
		if s.Error != "" {
			description += "Last poll: " + Clean(s.Error)
		}
		if err = form(ctx, huh.NewSelect[string]().Title("AZ VOTE MANAGER").Description(description).Options(options...).Value(&action)); err != nil {
			return err
		}
		switch action {
		case "exit":
			return nil
		case "projects", "phases":
			err = votables(ctx, c, s, action == "phases")
		case "follow":
			err = follow(ctx, c, s)
		case "activity":
			err = activity(ctx, s)
		case "mode":
			title := "Enable live voting for projects AND phase payouts?"
			if s.Live {
				title = "Pause all new votes? Already submitted transactions may still finish."
			}
			var ok bool
			ok, err = confirm(ctx, title)
			if err == nil && ok {
				_, err = control.Call(ctx, c, &control.Request{Action: "mode", Live: !s.Live})
			}
		}
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				continue
			}
			if err = note(ctx, "Action not completed", Clean(err.Error())); err != nil {
				return err
			}
		}
	}
	return nil
}
func managed(ctx context.Context, s *control.Snapshot) (string, error) {
	if len(s.Delegations) == 1 {
		return s.Delegations[0].Pillar, nil
	}
	options := []huh.Option[string]{}
	for _, d := range s.Delegations {
		options = append(options, huh.NewOption(Clean(d.Pillar), d.Pillar))
	}
	value := ""
	err := form(ctx, huh.NewSelect[string]().Title("Your pillar").Options(options...).Value(&value))
	return value, err
}
func follow(ctx context.Context, c config.Config, s *control.Snapshot) error {
	pillar, err := managed(ctx, s)
	if err != nil {
		return err
	}
	res, err := control.Call(ctx, c, &control.Request{Action: "pillars"})
	if err != nil {
		return err
	}
	options := []huh.Option[string]{}
	for _, name := range res.Pillars {
		if name != pillar {
			options = append(options, huh.NewOption(Clean(name), name))
		}
	}
	if len(options) == 0 {
		return errors.New("no other active pillars")
	}
	chosen := ""
	if err = form(ctx, huh.NewSelect[string]().Title("Copy votes from").Description("Type to search. Applies to project proposals and phase-payout requests; existing manual overrides stay in place.").Filtering(true).Options(options...).Value(&chosen)); err != nil {
		return err
	}
	yes, err := confirm(ctx, fmt.Sprintf("Have %s copy %s's votes?", Clean(pillar), Clean(chosen)))
	if err != nil || !yes {
		return err
	}
	_, err = control.Call(ctx, c, &control.Request{Action: "follow", Pillar: pillar, Follow: chosen})
	return err
}
func votables(ctx context.Context, c config.Config, s *control.Snapshot, phases bool) error {
	options := []huh.Option[int]{}
	for i, r := range s.Rows {
		if (r.Votable.Kind == "phase") == phases {
			label := Clean(r.Votable.Name)
			if phases {
				label = Clean(r.Votable.ProjectName) + " / " + label
			}
			options = append(options, huh.NewOption(fmt.Sprintf("%s  ·  %s  ·  you: %s / delegate: %s", label, Clean(r.Pillar), delegate.VoteText(r.Owner), delegate.VoteText(r.Delegate)), i))
		}
	}
	title := "Project proposals"
	if phases {
		title = "Phase payouts / payment requests"
	}
	if len(options) == 0 {
		return note(ctx, title, "No open requests in the latest snapshot. Refresh to check again. No vote can be cast on a closed or superseded request.")
	}
	index := 0
	if err := form(ctx, huh.NewSelect[int]().Title(title).Options(options...).Value(&index)); err != nil {
		return err
	}
	r := s.Rows[index]
	v := r.Votable
	details := fmt.Sprintf("Project: %s\nRequest: %s\nRequested: %s ZNN + %s QSR\nYour vote: %s    Delegate: %s\nStatus: %s\nID: %s\n\n%s\n\n%s", Clean(v.ProjectName), Clean(v.Name), v.ZNN, v.QSR, delegate.VoteText(r.Owner), delegate.VoteText(r.Delegate), r.Action, v.ID, preview(v.Description, 600), preview(v.URL, 150))
	action := ""
	if err := form(ctx, huh.NewSelect[string]().Title(Clean(v.Name)).Description(details).Options(huh.NewOption("Vote YES manually", "yes"), huh.NewOption("Vote NO manually", "no"), huh.NewOption("Vote ABSTAIN manually", "abstain"), huh.NewOption("Resume copying for this request", "resume"), huh.NewOption("Back", "back")).Value(&action)); err != nil {
		return err
	}
	if action == "back" {
		return nil
	}
	if !s.Live {
		return note(ctx, "Dry-run mode", "Enable live voting from the main menu before casting or resuming votes.")
	}
	message := fmt.Sprintf("Vote %s on %s? This stops copy voting for this request, even when your choice matches the existing vote.", strings.ToUpper(action), Clean(v.Name))
	if action == "resume" {
		message = "Remove the manual override and resume copying for this request?"
	}
	yes, err := confirm(ctx, message)
	if err != nil || !yes {
		return err
	}
	req := &control.Request{Action: "vote", Pillar: r.Pillar, ID: v.ID, Vote: action}
	if action == "resume" {
		req.Action = "resume"
	}
	res, err := control.Call(ctx, c, req)
	if err != nil {
		return err
	}
	return note(ctx, "Choice saved", res.Message+"\nUse Pending transactions & manual overrides to check progress.")
}
func activity(ctx context.Context, s *control.Snapshot) error {
	lines := []string{}
	if s.State != nil {
		for pillar, entries := range s.State.Pillars {
			for id, e := range entries {
				if e.Pending != nil {
					lines = append(lines, fmt.Sprintf("%s · %s\nPending %s · block %s", Clean(pillar), id, e.Pending.Vote.String(), e.Pending.BlockHash))
				}
				if e.ManualVote != nil {
					status := "queued"
					if e.ManualConfirmed {
						status = "confirmed"
					}
					lines = append(lines, fmt.Sprintf("%s · %s\nManual %s · %s", Clean(pillar), id, e.ManualVote.String(), status))
				} else if e.OwnerOverride {
					lines = append(lines, fmt.Sprintf("%s · %s\nExternal manual override", Clean(pillar), id))
				}
			}
		}
	}
	sort.Strings(lines)
	body := strings.Join(lines, "\n\n")
	if body == "" {
		body = "No pending submissions or manual overrides."
	}
	return note(ctx, "Activity", body)
}

func Wizard(ctx context.Context, path string) error {
	if !term.IsTerminal(os.Stdin.Fd()) {
		return errors.New("setup requires a terminal")
	}
	if _, err := os.Stat(path); err == nil {
		return errors.New("configuration already exists; use the menu to change its delegate")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	endpoint := "ws://127.0.0.1:35998"
	chain := "1"
	source := "znnd"
	if err = form(ctx, huh.NewInput().Title("Your Zenon node URL").Value(&endpoint).Validate(required), huh.NewInput().Title("Expected chain ID (mainnet: 1)").Value(&chain).Validate(func(s string) error {
		v, e := strconv.ParseUint(s, 10, 64)
		if e != nil || v == 0 {
			return errors.New("enter a positive chain ID")
		}
		return nil
	}), huh.NewSelect[string]().Title("Producer key").Options(huh.NewOption("Use existing znnd / NOMCTL configuration", "znnd"), huh.NewOption("Choose a key file and enter its password", "file")).Value(&source)); err != nil {
		return err
	}
	p := setup.Producer{KeyFilePath: filepath.Join(home, ".znn", "wallet", "producer")}
	if source == "znnd" {
		nodePath := filepath.Join(home, ".znn", "config.json")
		if err = form(ctx, huh.NewInput().Title("znnd config.json path").Description("Reads the producer key location and password locally. The original node config is never changed.").Value(&nodePath).Validate(absolute)); err != nil {
			return err
		}
		p, err = setup.ReadProducer(nodePath)
		if err != nil {
			return err
		}
	} else {
		index := "0"
		if err = form(ctx, huh.NewInput().Title("Producer key file").Value(&p.KeyFilePath).Validate(absolute), huh.NewInput().Title("Key index").Value(&index).Validate(func(s string) error {
			i, e := strconv.Atoi(s)
			if e != nil || i < 0 || i > 1000000 {
				return errors.New("invalid key index")
			}
			return nil
		}), huh.NewInput().Title("Producer key password").EchoMode(huh.EchoModePassword).Value(&p.Password).Validate(required)); err != nil {
			return err
		}
		p.Index, _ = strconv.Atoi(index)
	}
	id, _ := strconv.ParseUint(chain, 10, 64)
	d := config.Delegation{ProducerKeyFile: p.KeyFilePath, ProducerKeyIndex: p.Index}
	key, err := node.KeyFromFile(d, p.Password)
	if err != nil {
		return err
	}
	defer key.Destroy()
	address, err := key.GetAddress()
	if err != nil {
		return err
	}
	c := config.Config{NodeURL: endpoint, ChainID: id, PollIntervalSeconds: 30, PowTimeoutSeconds: 300}
	n, err := node.Open(c, nil, nil)
	if err != nil {
		return err
	}
	defer n.Close()
	pillars, err := n.Pillars(ctx)
	if err != nil {
		return err
	}
	options := []huh.Option[string]{}
	names := []string{}
	for name, info := range pillars {
		if info.RevokeTimestamp != 0 {
			continue
		}
		if info.ProducerAddress == *address {
			d.PillarName = name
		} else {
			names = append(names, name)
		}
	}
	if d.PillarName == "" {
		return errors.New("this producer key does not match an active pillar on that node")
	}
	sort.Strings(names)
	for _, name := range names {
		options = append(options, huh.NewOption(Clean(name), name))
	}
	if len(options) == 0 {
		return errors.New("no delegate pillars found")
	}
	if err = form(ctx, huh.NewSelect[string]().Title("Copy votes from").Description("Your pillar: "+Clean(d.PillarName)+". Type to search.").Filtering(true).Options(options...).Value(&d.Follow)); err != nil {
		return err
	}
	base, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return err
	}
	c.StatePath = filepath.Join(base, "state", "state.json")
	d.PasswordFile = filepath.Join(base, "producer-password")
	c.Delegations = []config.Delegation{d}
	n.Config = c
	n.Keys = node.Keys{d.PillarName: key}
	if err = n.Check(ctx); err != nil {
		return err
	}
	ok, err := confirm(ctx, "Save setup in dry-run mode? The producer password will be stored in an owner-only 0600 file for the background service.")
	if err != nil || !ok {
		if err == nil {
			return huh.ErrUserAborted
		}
		return err
	}
	if err = config.WritePrivate(d.PasswordFile, []byte(p.Password)); err != nil {
		return err
	}
	p.Password = ""
	if err = config.Save(path, c); err != nil {
		return err
	}
	return note(ctx, "Setup complete", "Open the menu to start the background service. It starts in dry-run mode; review project and phase-payout decisions before enabling live votes.")
}

func preview(s string, limit int) string {
	r := []rune(Clean(s))
	if len(r) > limit {
		return string(r[:limit]) + "…"
	}
	return string(r)
}
