Improves terminal-menu readability on dark backgrounds and terminals with customized colors.

- Explicit white text on a dark background for inputs, descriptions, choices, and navigation hints.
- Bright headings and selection indicators; inactive fields stay readable.
- High-contrast Yes/No buttons use explicit foreground/background pairs instead of the terminal's customizable ANSI palette.

Voting behavior and saved configuration are unchanged. Run the curl installer again to update, then reopen the menu. The theme fix applies to the menu immediately; existing v0.2.0 background services remain compatible.

**Compatibility:** This release supports the legacy fused-plasma / proof-of-work voting flow. Dynamic Plasma is not supported.
