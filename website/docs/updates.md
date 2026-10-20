---
title: Updates & upgrades
description: Upgrade from the terminal menu without losing settings or manual overrides.
---

## Upgrade from the menu

Open the app as its installation owner and select **Check for updates / upgrade**. If a newer stable version is available, the menu displays **Update available: v…** and offers **Upgrade application to v…**.

The upgrade screen shows the installed and available versions, executable path, release-notes link, and whether the existing service will restart. Confirm **Upgrade** to download and install.

The app verifies the download's SHA-256 checksum and checks the executable's reported version before replacing the existing binary. Configuration, passwords, voting mode, and saved manual overrides are preserved.

An active service restarts automatically only when its generated systemd unit is unmodified and matches this executable and configuration. A stopped service stays stopped. Custom units, units with drop-ins, and foreground processes need a manual restart; the upgrade screen explains this before confirmation. An upgrade does not install or enable another service.

Failed download or validation leaves the old executable in place. If installation succeeds but restarting the service fails, the app reports that partial result so you can check logs and restart it yourself.

## Upgrading to v0.3.0

Existing installations keep a zero-minute copy delay until you change it. After upgrading, choose **Copy vote delay: 0 minutes** to set your preferred delay. Existing manual overrides remain protected, and live timers persist on subsequent restarts.

The new configuration and state fields are not readable by older releases, which reject unknown fields. If you need the option to downgrade, keep a private pre-upgrade backup of configuration and state. Do not discard current pending transactions or manual overrides just to make an older binary start; resolve them before restoring older state.

## Update notices

The app checks GitHub when you open the menu and when you return to the main menu after 15 minutes. Selecting the update option always checks again. An idle menu updates when you return to or refresh the main screen.

These notices appear inside the menu. The daemon does not send notifications while it is closed. Failed update checks do not interrupt voting.

## Upgrade with curl

You can also rerun the installer as the **same user**, using the **same installation directory**. It replaces the executable in place and reconnects to your existing setup. For a root installation, these commands install without opening the menu, then restart the existing service:

```bash
curl -fsSL https://raw.githubusercontent.com/0x3639/az-vote-delegate/main/install.sh | sudo env AZ_NO_UI=1 sh
sudo systemctl restart az-vote-delegate.service
```

For a regular user installation:

```bash
curl -fsSL https://raw.githubusercontent.com/0x3639/az-vote-delegate/main/install.sh | AZ_NO_UI=1 sh
systemctl --user restart az-vote-delegate.service
```

The running service uses the old binary until restarted. Restarting uses the same named unit; it does not create a second service. If you run in the foreground, stop that process before starting the upgraded binary.

Changing users, configuration paths, or install directories creates a separate installation. Keep your original user and paths for normal upgrades.

If upgrading from v0.2.5 or earlier, use curl once to get the in-app upgrade option. See all [release notes on GitHub](https://github.com/0x3639/az-vote-delegate/releases).
