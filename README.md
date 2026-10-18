# az-vote-delegate

A terminal app and background service for Accelerator-Z voting. Choose a pillar
to copy, review project proposals and phase-payout requests, or cast your own
**yes / no / abstain** vote. Uses the existing pillar **producer key**, never the
owner's stake key. The menu works over SSH.

**[Documentation & quick start](https://0x3639.github.io/az-vote-delegate/)** —
installation, delegation, manual votes, and service management.

**Compatibility:** v0.2.7 supports the legacy fused-plasma / proof-of-work voting
flow. Dynamic Plasma is not supported by this release.

## TL;DR — quick start

On your Linux pillar host, install the app and open setup:

```sh
curl -fsSL https://raw.githubusercontent.com/0x3639/az-vote-delegate/main/install.sh | sudo sh
```

Use the user who owns your producer key; omit `sudo` if your node runs as an
unprivileged user.

1. **Set up:** confirm your node URL and chain ID, approve the background service,
   and select your existing znnd configuration or producer key file.
2. **Choose a pillar to copy:** pick from the shuffled top 30 voters, using the
   participation figures to help you choose.
3. **Choose a voting mode:** select **Dry run** to preview without submitting
   votes, or **Live** to copy that pillar's votes on AZ proposals and phase-payout
   requests. Confirm **Save and start** to save setup and start the service.

You can close the menu; the service keeps running. Reopen it anytime to change
your delegate or vote manually:

```sh
sudo az-vote-delegate
```

## Install

On a Linux amd64 or arm64 pillar host:

```sh
curl -fsSL https://raw.githubusercontent.com/0x3639/az-vote-delegate/main/install.sh | sudo sh
```

The installer automatically opens terminal setup, including when piped from curl.
Run `az-vote-delegate` (or `sudo az-vote-delegate`) anytime to reopen the menu.
Without a terminal, or with `AZ_NO_UI=1`, it installs the binary and prints the
command to run later.

Run the app as the user who owns your producer key. `sudo` is appropriate for a
root-operated znnd installation. For an unprivileged node user, install
without `sudo`; the binary goes to `~/.local/bin` and the installer prints its
full path. Root installation uses `/usr/local/bin`.

The installer downloads a prebuilt release, verifies its SHA-256 checksum and
replaces the executable atomically. The download step does not change node
settings or votes. The setup menu asks before saving configuration and installing
the service. Release binaries are static: no Go compiler, C compiler, Docker,
`gum` or `jq` is required on the pillar host.

To pin a release:

```sh
curl -fsSL https://raw.githubusercontent.com/0x3639/az-vote-delegate/main/install.sh | sudo env AZ_VERSION=v0.2.7 sh
```

Run the installer again as the **same user**, using the **same install directory**,
to upgrade. It detects the existing executable and replaces it in place; it does
not create or start an additional service. Your configuration, voting mode and
saved manual overrides stay in place. The menu reconnects to the existing service
when it finds your configuration, instead of running first-time setup again.

The running service keeps using its old executable until you restart it. For a
root installation:

```sh
sudo systemctl restart az-vote-delegate.service
```

For a user installation, use `systemctl --user restart az-vote-delegate.service`.
Restart replaces the running process for the same named service. Repeating
service installation also uses that same unit name. A state-file lock prevents
two workers from managing the same state at once. If you run in the foreground,
stop that process before launching the upgraded version.

Use `AZ_INSTALL_DIR=/absolute/path` to choose another binary directory. Switching
users, configuration files or install directories is a separate installation,
not an upgrade; keep using the original user and paths for normal reinstalls.

## Updates from the menu

Choose **Check for updates / upgrade** from the main menu. When a newer stable
release is available, the menu shows **Update available: v…** and the option
becomes **Upgrade application to v…**. The app checks GitHub when you open the
menu and when you return to the main menu after 15 minutes. Selecting the update
option always checks again. These are in-app notices; the daemon does not send
notifications while the menu is closed, and an idle screen updates when you
return to or refresh the main menu. Failed checks do not interrupt voting.

The upgrade screen shows both versions, the installation path, a release-notes
link and whether the existing service will restart. Confirm **Upgrade** to:

1. Download the Linux amd64 or arm64 release, verify its SHA-256 checksum, and
   check that the executable reports the expected version.
2. Replace the existing executable in place. Configuration, passwords, voting
   mode and persistent manual overrides remain unchanged.
3. Restart the existing active systemd service when its unmodified generated
   unit matches this executable and configuration, then reopen the updated menu.

A stopped service stays stopped. Foreground processes, custom units and units
with drop-ins require a manual restart; the screen explains this before you
confirm. No upgrade installs or enables a second service. If a service restart
fails after the binary was installed, the result explicitly reports that partial
outcome so you can inspect the service logs and restart it manually.

Run the menu as the installation owner (`sudo az-vote-delegate` for a root
installation). The app does not elevate privileges automatically. Downloads and
validation happen while the existing service keeps running; only the service
restart briefly interrupts it. Failed download, checksum or version validation
leaves the old executable in place. Downgrades and concurrent upgrades are
rejected.

**First upgrade from v0.2.5 or earlier:** rerun the curl installer once to get
this menu option, then restart your existing service. Future upgrades can use
the menu.

## First run

The setup wizard guides you through:

1. Your local/trusted node URL and expected chain ID (mainnet is `1`).
2. Confirming whether to install and start the background service after setup.
3. Importing the producer settings from your existing znnd `config.json`, or
   entering a key-file path, derivation index and password. No key is generated.
4. Identifying your pillar and choosing a delegate from the top 30 active voters,
   shuffled and shown with their voting participation.
5. Choosing **Dry run — preview without submitting votes** or **Live — submit
   votes**. Dry run is selected by default.
6. Confirming the selected mode and saving setup. The app saves a private
   configuration and password file, then installs and starts the service if you
   requested it. The service starts in the mode you selected.

Use **Voting mode: Dry run / Live** in the main menu to change modes later. Live
mode submits votes; dry run only previews decisions. Switching to dry run does
not undo transactions already submitted. The saved mode survives service
restarts. Closing the menu leaves the service running. The original znnd
configuration is never changed.

By default, application files live under `~/.config/az-vote-delegate/`:
`config.json`, `producer-password`, and a private `state/` directory. An existing
`./config.json` is preferred for compatibility; `--config` selects an explicit
path. Keep application files private and out of Git.

On root-operated hosts the generated systemd unit is installed in
`/etc/systemd/system`. Otherwise it is a systemd **user** unit. For a user service
to survive logout and start at boot, enable lingering for that node user:

```sh
sudo loginctl enable-linger YOUR_NODE_USER
```

If systemd is unavailable, run `az-vote-delegate run` in another terminal and
reconnect from the menu. The service and UI must use the same config and user.

## The menu

Press **Esc** or select **Back** to return one screen: vote confirmation →
request details → request list → main menu. Back cancels an unsubmitted choice.
At the main menu, Esc exits the menu; **Ctrl+C** exits from any screen. The
background service keeps running. During first-time setup, Esc cancels unsaved
setup; **Shift+Tab** moves to the previous field within a setup step.

When searching a list, **Tab** applies the filter and **Ctrl+R** clears it.
Esc still goes back while searching.

- **Vote on unanswered AZ proposals**: inspect and vote on open initial funding
  proposals your pillar has not voted on yet.
- **Vote on unanswered phase payouts**: inspect unpaid phase requests your pillar
  has not voted on, with their parent project, description, URL and requested
  ZNN/QSR. Copy voting and manual voting work for both kinds of request.
- **Choose pillar to copy**: choose from the shuffled top 30 active voters, with
  participation shown beside each name. Existing manual overrides stay in place.
- **Check for updates / upgrade**: check the latest stable GitHub release and
  install it after confirmation. Available updates appear in the menu.
- **Pending transactions & manual overrides**: inspect queued choices, confirmed
  manual overrides and pending block hashes, or resume copying for an override.
- **Voting mode: Dry run / Live**: explicitly select previews only or live vote
  submission, then confirm the change. Already submitted transactions may still
  finish after switching to dry run.

The manual lists fetch fresh votes from the node. Any recorded yes, no or abstain
hides that request; queued manual choices also leave the list and appear in
Activity. A final check rejects the selection if your pillar has voted since you
opened the list. A pending copy that has not executed can still be selected for
a manual override.

### Delegate participation

The picker ranks active pillars by the number of distinct requests they voted
on, excludes your own pillar, takes the top 30 (or all if fewer), then shuffles
the display each time it opens. Equal counts are randomly ordered before the
cutoff so ties do not favor alphabetically earlier names.

Each option shows `82.0% participation · 82/100 requests`, for example. Yes, no
and abstain all count. The denominator is all project proposals and retained
phase requests returned by the node, including closed requests. This is not
adjusted for when a pillar joined; deleted/replaced phase revisions are not
available through this API. With no recorded requests, the menu says so rather
than showing a fabricated percentage. Changing the delegate uses the same picker.

The app votes on payment requests; it does not submit a project owner's payment
request or promise that voting will immediately release funds. Contract rules
and the votes of other pillars still determine whether a phase passes.

## Manual votes take priority

A manual vote affects **only that proposal or phase-payout request**. The service
keeps running and copying all other eligible requests. There is no pause for a
number of days and no timed expiry of your manual choice.

A manual vote from this app records an explicit, durable override **before** it
is acknowledged. A manual yes remains a manual yes even if the delegate later
changes to no. The manual lists show unanswered requests; the underlying override
tracking also distinguishes an explicit same-value choice from a copied vote.
Overrides are stored locally by **pillar name + request ID** in the configured
`statePath` file (normally `~/.config/az-vote-delegate/state/state.json`). Queued
and confirmed manual choices survive service restarts, host reboots and binary
upgrades. Keep this file when migrating the application.

The service is the only signer and state writer; the UI uses its owner-only Unix
socket and never starts a competing signer.

When you choose a manual vote, the service interrupts copy-vote preparation. If
a copy vote has already been submitted, the app waits for its confirmed contract
receipt before applying your choice. The menu distinguishes a **queued choice**
from a **confirmed manual vote**. A lost publish response is held pending rather
than blindly retried. You may need to inspect that transaction on your trusted
node before proceeding. If the request closes while an earlier transaction is
pending, your queued choice cannot reopen it or undo the contract's decision.

The manual override remains for that request until you choose **Resume copying**
in Activity, the request closes, or its confirmed votes are reset. Updating
a phase in go-zenon replaces its votable ID; the revised payout request gets its
own fresh voting round. The app follows the current phase ID and never applies
an old phase's override to its replacement. Projects and phases have independent
IDs and overrides.

Votes cast through **external tools** still have the off-chain limitation: the
current vote value cannot identify a same-value manual vote, and an external
writer can race with this service. For guaranteed external manual control, stop
the service, resolve pending submissions, and leave it stopped while voting.

## Commands and configuration

```sh
az-vote-delegate                         # menu; setup on first use
az-vote-delegate setup                   # setup wizard only
az-vote-delegate service-install         # install/start the systemd service
az-vote-delegate run                     # service, using the saved voting mode
az-vote-delegate run --dry-run           # force previews only
az-vote-delegate run --dry-run=false     # explicitly enable live submissions
az-vote-delegate status                  # current project/phase decisions
az-vote-delegate check                   # key, pillar and chain-ID checks
az-vote-delegate version
```

Every relevant command accepts `--config /absolute/path/config.json`; `run`
also accepts `--debug`. `check` never signs or publishes. If the service is
running, `status` reads its cached snapshot without opening another key.

[config.example.json](config.example.json) supports multiple delegations, one
per producer. Paths resolve relative to the config. `live` defaults to false;
`pollIntervalSeconds` defaults to 30 and `powTimeoutSeconds` to 300. The chain ID
must be explicit. Configure exactly one of `passwordEnv` or `passwordFile` for
each key. Password files must be regular files with mode 0600. Passwords are not
accepted as CLI flags or stored in the app's JSON config.

The wizard uses a password file so the background service can unlock the key
after reboot. For an environment-based configuration, arrange the environment
in your systemd unit, for example using the example in
[deploy/az-vote-delegate.service](deploy/az-vote-delegate.service).

## Reliability and operation

The service reads current votes every poll, so it does not need to observe the
delegate's original vote block. It pages through all projects, using the node's
1,024-entry limit, and selects Voting projects within the 14-day window and the
current Voting phase of Active projects. Reads use node time. Delegation chains
work; an empty A↔B cycle produces no initial vote.

The state file tracks copy votes, explicit manual choices, external overrides,
backoff and pending submissions. Saves use a 0600 temporary file, file sync,
atomic rename and directory sync. Save failures stop live operation. A file lock
prevents a second service from using the same state, and the private directory
and mode-0600 control socket restrict control to its owner. Do not run multiple
instances for the same producer using different state paths.

A signed block hash is journaled before publication. Pending sends wait for a
confirmed contract receive block, then votes are reread. Preparation failures
and definite rejections retry on the next poll, then back off exponentially up
to 30 minutes. Ambiguous publications remain pending, including a crash between
journaling and publishing. This favors avoiding duplicate votes over automatic
recovery from an unknown outcome.

Missing state conservatively treats existing votes as manual. Corrupt or
cross-network state is refused. After stopping the service and resolving every
pending transaction, an explicit reset can discard attribution and overrides:

```sh
az-vote-delegate run --config /path/to/config.json --dry-run=false --reset-state
```

Existing on-chain votes remain. Resetting while a transaction might still execute
is unsafe; back up and inspect the state first.

The SDK asks the node whether fused plasma is sufficient, otherwise it computes
PoW. The service does not auto-fuse QSR. Logs identify PoW versus fused plasma.
Reading the node's plasma requirements does not establish Dynamic Plasma
compatibility; that requires a separate implementation and validation effort.
Sends are sequential, so CPU work can extend a poll beyond the configured delay.
A status preview remains available while PoW runs. SIGINT/SIGTERM cancel work
and save state. WebSocket reads are interrupted on cancellation; HTTP reads have
a 30-second transport timeout. Subsequent polls open a fresh connection.

The producer key is held in memory while running, as it is in znnd. Use the node
host or a host with equivalent protection, and an RPC endpoint you trust. No
browser server, remote control port or hosted key service is involved.

## Development and releases

Source builds need Go 1.26 (toolchain 1.26.7) and a C compiler, because the SDK's
transitive go-zenon/secp256k1 dependency requires CGO. macOS is supported for
source-based development; prebuilt releases target Linux amd64 and arm64.

```sh
make build
make test       # Go race detector, local mock RPC servers
make vet
make lint       # golangci-lint v2.13.1 or compatible
sh scripts/test-install.sh
```

The Makefile disables a parent `go.work`. Tests cover reconciliation, native
znnd wallet compatibility, signed embedded votes and PoW, project/phase paging,
manual same-value overrides, restart recovery, interruption of copy preparation,
waiting for pending copies, phase-payout replacement, participation ranking and
shuffling, unanswered-request filtering, and installer checksum failure. The
installer tests also exercise a piped installation in a pseudo-terminal to verify
automatic menu launch and its opt-out. No operator credentials are used in tests.

Pushing a version tag runs [the release workflow](.github/workflows/release.yml):
tests and lint on Linux, static CGO builds for amd64 and arm64, smoke tests on
both (arm64 via QEMU), archive/checksum creation and a draft GitHub release.
Publish that draft only after the build and asset checks pass.

Live mainnet/testnet acceptance remains separate. Before operating with real
votes, run `check` and review dry-run output on the intended node. On an isolated
testnet, verify project yes→no copying, same-value manual takeover, a manual vote
while copying, phase-payout copying/manual override and a phase revision. No
live voting acceptance is implied by the offline tests.

## License

GPL-3.0; see [LICENSE](LICENSE). Release archives also include dependency
license and notice files, including the SDK and go-zenon licenses.
