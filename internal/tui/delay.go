package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/0x3639/az-vote-delegate/internal/config"
	"github.com/0x3639/az-vote-delegate/internal/control"
	"github.com/charmbracelet/huh"
)

func parseDelay(value string) (int, error) {
	minutes, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("enter a whole number of minutes (0..10080)")
	}
	if err = config.ValidateCopyDelay(minutes); err != nil {
		return 0, err
	}
	return minutes, nil
}

func chooseDelay(ctx context.Context, current int) (int, error) {
	value := strconv.Itoa(current)
	err := form(ctx, huh.NewInput().Title("Copy vote delay (minutes)").Description("Wait after observing each vote before copying it. Applies to AZ proposals and phase payouts; manual votes are immediate. 0 disables the delay. Votes remain public; a delay does not guarantee privacy.").Value(&value).Validate(func(s string) error { _, err := parseDelay(s); return err }))
	if err != nil {
		return current, err
	}
	return parseDelay(value)
}

func changeDelay(ctx context.Context, c config.Config, current int) error {
	minutes, err := chooseDelay(ctx, current)
	if err != nil || minutes == current {
		return err
	}
	_, err = control.Call(ctx, c, &control.Request{Action: "delay", CopyDelayMinutes: minutes})
	return err
}
