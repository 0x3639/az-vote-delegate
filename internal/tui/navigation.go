package tui

import (
	"context"
	"errors"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

var errBack = errors.New("back")

// Keep form rendering separate from menu transitions so navigation can be
// exercised with keyboard input without a real terminal or producer key.
var runForm = interactiveForm

func navigationForm(fields ...huh.Field) (*huh.Form, *bool) {
	back := false
	keys := huh.NewDefaultKeyMap()
	// Esc always leaves the current screen, including while filtering.
	keys.Select.SetFilter.SetKeys("tab")
	keys.Select.SetFilter.SetHelp("tab", "apply filter")
	keys.Select.ClearFilter.SetKeys("ctrl+r")
	keys.Select.ClearFilter.SetHelp("ctrl+r", "clear filter")
	f := huh.NewForm(huh.NewGroup(fields...).Description("Esc: back / cancel · Ctrl+C: exit menu")).WithTheme(menuTheme()).WithKeyMap(keys)
	f.WithProgramOptions(tea.WithOutput(os.Stderr), tea.WithReportFocus(), tea.WithFilter(func(_ tea.Model, msg tea.Msg) tea.Msg {
		if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEsc {
			back = true
			return tea.KeyMsg{Type: tea.KeyCtrlC}
		}
		return msg
	}))
	return f, &back
}

func formResult(err error, back bool) error {
	if back && errors.Is(err, huh.ErrUserAborted) {
		return errBack
	}
	return err
}

func interactiveForm(ctx context.Context, fields ...huh.Field) error {
	f, back := navigationForm(fields...)
	err := f.RunWithContext(ctx)
	return formResult(err, *back)
}
