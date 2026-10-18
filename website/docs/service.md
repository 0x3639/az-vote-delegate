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
| `config.json` | Node, key, delegation, and voting-mode settings |
| `producer-password` | Owner-only password file created by setup |
| `state/state.json` | Manual overrides, copied votes, and pending submissions |

An existing `./config.json` is preferred for compatibility. Use `--config /absolute/path/config.json` to select a specific configuration. The menu and service must use the same file.

Restarts and upgrades preserve manual choices. When migrating, stop the service and preserve the configuration, password, key, and state together, keeping them private. Resolve pending transactions before changing the setup. Do not delete state to clear an error: it contains the record of submissions and manual overrides.

A state-file lock prevents two workers from using the same state. Do not create separate state files for competing instances with the same producer.
