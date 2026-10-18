package tui

import (
	"context"
	"errors"

	"github.com/0x3639/az-vote-delegate/internal/config"
	"github.com/0x3639/az-vote-delegate/internal/control"
	"github.com/charmbracelet/huh"
)

func modeName(live bool) string {
	if live {
		return "Live"
	}
	return "Dry run"
}

func modeDescription(live bool) string {
	if live {
		return "Live: the service submits votes on AZ proposals and phase payouts."
	}
	return "Dry run: preview decisions without submitting votes."
}

func chooseMode(ctx context.Context, current bool, backLabel string) (bool, error) {
	selected := 0
	if current {
		selected = 1
	}
	err := form(ctx, huh.NewSelect[int]().Title("Voting mode").Description("Choose how the service handles AZ proposals and phase payouts. You can change this later from the main menu.").Options(
		huh.NewOption("Dry run — preview without submitting votes", 0),
		huh.NewOption("Live — submit votes", 1),
		huh.NewOption(backLabel, -1),
	).Value(&selected))
	if err == nil && selected == -1 {
		return current, errBack
	}
	return selected == 1, err
}

func changeMode(ctx context.Context, c config.Config, current bool) error {
	for {
		live, err := chooseMode(ctx, current, "Back")
		if err != nil || live == current {
			return err
		}
		description := modeDescription(live)
		if !live {
			description += " Already submitted transactions may still finish."
		}
		ok := false
		err = form(ctx, huh.NewConfirm().Title("Switch to "+modeName(live)+" mode?").Description(description).Affirmative("Switch mode").Negative("Back").Value(&ok))
		if errors.Is(err, errBack) || (err == nil && !ok) {
			continue
		}
		if err != nil {
			return err
		}
		_, err = control.Call(ctx, c, &control.Request{Action: "mode", Live: live})
		return err
	}
}
