package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func menuTheme() *huh.Theme {
	t := huh.ThemeBase()
	// Leave backgrounds transparent: nested Huh/Lipgloss views reset ANSI
	// styles between text and padding, so painted backgrounds form patchwork.
	// Use explicit palette colors for readable text on dark and light terminals.
	text := lipgloss.AdaptiveColor{Light: "235", Dark: "255"}
	accent := lipgloss.AdaptiveColor{Light: "22", Dark: "120"}
	errorText := lipgloss.AdaptiveColor{Light: "124", Dark: "210"}
	buttonText := lipgloss.Color("255")
	normal := lipgloss.NewStyle().Foreground(text)
	t.Form.Base = normal
	t.Group.Base = normal
	for _, f := range []*huh.FieldStyles{&t.Focused, &t.Blurred} {
		for _, style := range []*lipgloss.Style{
			&f.Base, &f.Card, &f.Title, &f.NoteTitle, &f.Description,
			&f.Option, &f.SelectedOption, &f.UnselectedOption,
			&f.SelectSelector, &f.MultiSelectSelector, &f.SelectedPrefix, &f.UnselectedPrefix,
			&f.NextIndicator, &f.PrevIndicator, &f.Directory, &f.File,
			&f.TextInput.Text, &f.TextInput.Prompt, &f.TextInput.Placeholder,
			&f.ErrorIndicator, &f.ErrorMessage,
		} {
			*style = style.Foreground(text)
		}
		f.Title = f.Title.Bold(true)
		f.NoteTitle = f.NoteTitle.Bold(true)
		f.SelectSelector = f.SelectSelector.Foreground(accent).Bold(true)
		f.MultiSelectSelector = f.MultiSelectSelector.Foreground(accent).Bold(true)
		f.SelectedPrefix = f.SelectedPrefix.Foreground(accent).Bold(true)
		f.TextInput.Cursor = normal
		f.TextInput.CursorText = normal
		f.ErrorIndicator = f.ErrorIndicator.Foreground(errorText).Bold(true)
		f.ErrorMessage = f.ErrorMessage.Foreground(errorText).Bold(true)
		f.FocusedButton = f.FocusedButton.Foreground(lipgloss.Color("232")).Background(buttonText).Bold(true)
		f.BlurredButton = f.BlurredButton.Foreground(buttonText).Background(lipgloss.Color("238"))
		f.Next = f.FocusedButton
	}
	t.Focused.Base = t.Focused.Base.BorderForeground(accent)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(accent)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(accent)
	t.Group.Title = t.Focused.Title
	t.Group.Description = normal
	// Navigation hints must not disappear into the background either.
	t.Help.ShortKey, t.Help.FullKey = normal.Bold(true), normal.Bold(true)
	t.Help.ShortDesc, t.Help.FullDesc = normal, normal
	t.Help.ShortSeparator, t.Help.FullSeparator, t.Help.Ellipsis = normal, normal, normal
	return t
}
