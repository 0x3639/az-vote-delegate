Adds application updates to the terminal menu.

- Displays a notice when a newer stable GitHub release is available, with installed and available versions. Checks run when opening the menu and on returning after 15 minutes; the menu option always checks again. No automatic installation occurs.
- **Check for updates / upgrade** downloads and verifies the matching Linux release, checks the executable version, and atomically replaces the existing binary after confirmation.
- Preserves configuration, voting mode, passwords and persistent manual overrides. Downloads happen while voting continues.
- Restarts only the existing active systemd service with a matching generated unit, then reopens the updated menu. Stopped services stay stopped; foreground processes and custom units receive manual restart instructions. No additional service is created.
- Rejects invalid checksums, wrong versions, downgrades and concurrent upgrades. Reports an installed binary separately from a failed service restart.

Use the curl installer once to upgrade from v0.2.5 or earlier. Subsequent upgrades can use the menu. Update notices appear inside the app, not as notifications while the menu is closed.

Validated offline with HTTP fixtures, temporary executables, keyboard navigation and mocked service commands. Linux amd64 and arm64 binaries and the installer are checked by the release workflow.

**Compatibility:** This release supports the legacy fused-plasma / proof-of-work voting flow. Dynamic Plasma is not supported.
