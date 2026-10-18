The curl installer now opens terminal setup automatically. Setup asks for confirmation to install and start the background service, then lets the operator select a delegate.

**Compatibility:** This release supports the legacy fused-plasma / proof-of-work voting flow. Dynamic Plasma is not supported.

- The delegate picker takes the top 30 active pillars by recorded AZ votes, excludes the operator's own pillar, and shuffles the display. Ties are shuffled before the cutoff.
- Each pillar shows participation as a percentage and voted/total request counts. Counts include all project and phase requests retained by the node, including closed requests; they are not adjusted for pillar registration dates.
- Manual project and phase-payout menus fetch fresh votes and show only unanswered requests. Yes, no and abstain all count as answered. Queued manual choices appear in Activity.
- The service rechecks an unanswered request before accepting a manual vote. Activity retains the ability to resume copying after a manual override.
- `AZ_NO_UI=1` skips automatic menu launch for scripted installations; installations without a terminal print the command to run later.

The service starts in dry-run mode. Enable live voting from the menu after reviewing setup. Existing installations should restart their voting service after upgrading.

Validated with offline RPC fixtures, race tests, installer pseudo-terminal tests, checksum checks, and Linux amd64/arm64 binary startup tests. No live-chain votes or operator keys were used.
