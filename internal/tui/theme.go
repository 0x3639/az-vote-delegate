package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func menuTheme() *huh.Theme {
	t := huh.ThemeBase()
	// Explicit foreground/background pairs avoid inheriting a terminal's custom
	// default text color or its remapped ANSI 0–15 palette. Keep inactive fields
	// readable too: the setup wizard displays several fields at once.
	background := lipgloss.Color("234")
	text := lipgloss.Color("255")
	accent := lipgloss.Color("120")
	normal := lipgloss.NewStyle().Foreground(text).Background(background)
	t.Form.Base = normal
	t.Group.Base = normal
	t.FieldSeparator = t.FieldSeparator.Background(background)
	for _, f := range []*huh.FieldStyles{&t.Focused, &t.Blurred} {
		for _, style := range []*lipgloss.Style{
			&f.Base, &f.Card, &f.Title, &f.NoteTitle, &f.Description,
			&f.Option, &f.SelectedOption, &f.UnselectedOption,
			&f.SelectSelector, &f.MultiSelectSelector, &f.SelectedPrefix, &f.UnselectedPrefix,
			&f.NextIndicator, &f.PrevIndicator, &f.Directory, &f.File,
			&f.TextInput.Text, &f.TextInput.Prompt, &f.TextInput.Placeholder,
			&f.ErrorIndicator, &f.ErrorMessage,
		} {
			*style = style.Foreground(text).Background(background)
		}
		f.Title = f.Title.Bold(true)
		f.NoteTitle = f.NoteTitle.Bold(true)
		f.SelectSelector = f.SelectSelector.Foreground(accent).Bold(true)
		f.MultiSelectSelector = f.MultiSelectSelector.Foreground(accent).Bold(true)
		f.SelectedPrefix = f.SelectedPrefix.Foreground(accent).Bold(true)
		f.TextInput.Cursor = normal
		f.TextInput.CursorText = normal
		f.ErrorIndicator = f.ErrorIndicator.Foreground(lipgloss.Color("210")).Bold(true)
		f.ErrorMessage = f.ErrorMessage.Foreground(lipgloss.Color("210")).Bold(true)
		f.FocusedButton = f.FocusedButton.Foreground(lipgloss.Color("232")).Background(text).Bold(true)
		f.BlurredButton = f.BlurredButton.Foreground(text).Background(lipgloss.Color("238"))
		f.Next = f.FocusedButton
	}
	t.Focused.Base = t.Focused.Base.BorderForeground(accent).BorderBackground(background)
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
