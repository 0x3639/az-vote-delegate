Adds Back navigation throughout the terminal menu.

- Press **Esc** or choose **Back** to leave a submenu without closing the app.
- Back returns from vote confirmation to request details, then to the unanswered list, then to the main menu. Cancelling does not submit a vote or change voting mode.
- Delegate selection, voting mode, and manual-override confirmations also return to their previous screen.
- Every screen displays the Esc / Ctrl+C shortcut hint. Esc also works while searching; Tab applies a filter and Ctrl+R clears it.
- Esc at the main menu and Ctrl+C exit the menu while the background service keeps running. Esc during first-time setup cancels unsaved setup.

Existing settings and votes are preserved. Back does not undo an already submitted vote.

**Compatibility:** This release supports the legacy fused-plasma / proof-of-work voting flow. Dynamic Plasma is not supported.
