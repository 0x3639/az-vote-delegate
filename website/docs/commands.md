---
title: Command reference
description: Terminal commands and configuration for az-vote-delegate.
---

Run commands as the installation owner. Use `sudo` for root installations, or `~/.local/bin/az-vote-delegate` if a regular user's install directory is not on `PATH`.

| Command | Purpose |
| --- | --- |
| `az-vote-delegate` | Open the menu; start setup on first use |
| `az-vote-delegate setup` | Open the setup wizard |
| `az-vote-delegate service-install` | Install and start the systemd service |
| `az-vote-delegate run` | Run the worker using the saved voting mode |
| `az-vote-delegate run --dry-run` | Force previews without submitting votes |
| `az-vote-delegate run --dry-run=false` | Explicitly enable live submissions |
| `az-vote-delegate status` | Show current project and phase decisions |
| `az-vote-delegate check` | Check the key, pillar, and chain ID without signing or publishing |
| `az-vote-delegate version` | Print the installed version |

The menu is the interface for selecting delegates, casting manual votes, and upgrading. Do not launch `run` alongside an existing service.

## Select a configuration

Relevant commands accept an explicit path:

```bash
az-vote-delegate status --config /absolute/path/config.json
```

If the service is running, `status` reads its cached snapshot without opening a second key. `run` also accepts `--debug` for additional logging.

The default path is `~/.config/az-vote-delegate/config.json`, though an existing `./config.json` takes precedence. Paths within the configuration resolve relative to the configuration file.

## Configuration fields

See the complete [example configuration](https://github.com/0x3639/az-vote-delegate/blob/main/config.example.json).

| Setting | Behavior |
| --- | --- |
| `live` | Defaults to `false` (Dry run) |
| `pollIntervalSeconds` | Defaults to `30` |
| `copyDelayMinutes` | Defaults to `0`; whole minutes from `0` to `10080`. Applies to copied project and phase votes, not manual votes. |
| `powTimeoutSeconds` | Defaults to `300` |
| Chain ID | Must be explicitly configured |
| `passwordEnv` / `passwordFile` | Configure exactly one per key |
| `statePath` | Persistent voting state, including manual overrides and live copy-delay observations |

Password files must be regular files with mode `0600`. Passwords are not accepted as CLI flags or stored directly in the JSON configuration. Setup uses a password file so the background service can unlock the producer key after reboot.

Advanced configurations can include multiple delegations, one per producer. If you use an environment variable for a password, supply it to the service environment; see the [example systemd unit](https://github.com/0x3639/az-vote-delegate/blob/main/deploy/az-vote-delegate.service).

## Installer options

| Environment variable | Purpose |
| --- | --- |
| `AZ_NO_UI=1` | Install without opening the terminal menu |
| `AZ_VERSION=v…` | Install a specific release tag |
| `AZ_INSTALL_DIR=/absolute/path` | Set the binary installation directory |

When upgrading, keep the original user and installation directory. See [Updates & upgrades](./updates.md).

Change the delay through the menu to apply it immediately. If you edit JSON directly, restart the existing service to load the new configuration. For example, add `"copyDelayMinutes": 15` at the top level of `config.json` for a 15-minute delay. See [delay behavior and limitations](./delegation.md#delay-copied-votes).
