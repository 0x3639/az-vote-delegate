Add a configurable delay before copying votes on AZ proposals and phase-payout requests.

- Set **Copy vote delay** during setup or from the main menu. Choose 0–10,080 whole minutes; 0 preserves immediate copying on existing installations.
- Each request waits independently from when the service observes a vote that needs copying. Live-mode timers survive restarts. A changed delegate vote starts a new timer.
- Manual votes bypass the delay and retain their per-request overrides. Other requests continue normally.
- View waiting votes and their earliest submission times in **Pending transactions & manual overrides**. Settings apply without restarting the service.
- The README is now a short quick start with links to the full documentation.

Submission occurs on a poll after the delay, subject to node availability, retries and proof of work. Requests may close before a delayed vote can be submitted. A delay does not conceal public voting patterns or guarantee privacy.

Upgrade through **Check for updates / upgrade**, or rerun the curl installer using the original installation account. Existing configurations default to zero delay. Dry-run observations stay in memory; enabling Live reloads durable state. Config/state files written by this release contain new fields that older versions cannot read; keep a pre-upgrade backup if you need to downgrade, and preserve manual overrides.

**Compatibility:** Legacy fused plasma / proof of work only. Dynamic Plasma is not supported.
