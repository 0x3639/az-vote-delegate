---
title: Background service
description: Check the voting service, read logs, and preserve its state.
---

The setup wizard can install and start a systemd service. The terminal menu controls that service; closing the menu leaves it running.

## Root installation

```bash
sudo systemctl status az-vote-delegate.service
sudo journalctl -u az-vote-delegate.service -n 100 --no-pager
```

To restart the existing service:

```bash
sudo systemctl restart az-vote-delegate.service
```

To stop it:

```bash
sudo systemctl stop az-vote-delegate.service
```

The generated unit is installed under `/etc/systemd/system`.

## Regular user installation

Run these as the same user that installed the application:

```bash
systemctl --user status az-vote-delegate.service
journalctl --user -u az-vote-delegate.service -n 100 --no-pager
systemctl --user restart az-vote-delegate.service
```

For the user service to survive logout and start at boot, enable lingering. Replace the placeholder with your node's account name:

```bash
sudo loginctl enable-linger YOUR_NODE_USER
```

Use `systemctl --user stop az-vote-delegate.service` to stop it.

## If you skipped service installation

After saving setup, run this as the installation owner to install and start the service:

```bash
az-vote-delegate service-install
```

For root installations, prefix the command with `sudo`. For user installations without `~/.local/bin` on `PATH`, use `~/.local/bin/az-vote-delegate`.

If systemd is unavailable, keep the worker running in a terminal:

```bash
az-vote-delegate run
```

Open the menu in another terminal, with the same user and configuration. `run` uses your saved Dry run or Live mode. Do not start a foreground worker alongside the service.

## Configuration and persistent state

By default, files are stored under `~/.config/az-vote-delegate/` for the user running the app:

| File | Purpose |
| --- | --- |
| `config.json` | Node, key, delegation, copy-delay, and voting-mode settings |
| `producer-password` | Owner-only password file created by setup |
| `state/state.json` | Manual overrides, copy-delay observations, copied votes, and pending submissions |

An existing `./config.json` is preferred for compatibility. Use `--config /absolute/path/config.json` to select a specific configuration. The menu and service must use the same file.

Restarts and upgrades preserve manual choices. When migrating, stop the service and preserve the configuration, password, key, and state together, keeping them private. Resolve pending transactions before changing the setup. Do not delete state to clear an error: it contains the record of submissions and manual overrides.

A state-file lock prevents two workers from using the same state. Do not create separate state files for competing instances with the same producer.

## Reliability details

The service reads current votes every poll, so it does not need to observe the
delegate's original vote block. It pages through all projects, using the node's
1,024-entry limit, and selects Voting projects within the 14-day window and the
current Voting phase of Active projects. Reads use node time. Delegation chains
work; an empty A↔B cycle produces no initial vote.

The state file tracks copy votes, explicit manual choices, external overrides,
copy-delay observations, backoff and pending submissions. Saves use a 0600 temporary file, file sync,
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
Sends are sequential, so CPU work can extend a poll beyond the configured poll interval.
A status preview remains available while PoW runs. SIGINT/SIGTERM cancel work
and save state. WebSocket reads are interrupted on cancellation; HTTP reads have
a 30-second transport timeout. Subsequent polls open a fresh connection.

The producer key is held in memory while running, as it is in znnd. Use the node
host or a host with equivalent protection, and an RPC endpoint you trust. No
browser server, remote control port or hosted key service is involved.
