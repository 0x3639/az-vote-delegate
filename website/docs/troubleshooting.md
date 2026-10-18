---
title: Troubleshooting
description: Find the next step for connection, service, and voting problems.
---

## The app is running but does not vote

Check the mode shown at the top of the menu. **Dry run** only previews decisions. Select **Voting mode: Dry run / Live** to change it.

In Live mode, confirm the [background service is running](./service.md). Your delegate must have voted on an eligible request, and an existing manual override can prevent copying that request. Check **Pending transactions & manual overrides** and service logs.

The app supports legacy fused plasma and proof of work. Dynamic Plasma is not supported. Proof of work can take time; sends are sequential and can extend a poll beyond its configured delay.

## Cannot connect to the node

Verify that znnd is running and the configured RPC URL is reachable from the app's host. The default is `ws://127.0.0.1:35998`. Use your local or trusted node and the correct chain ID; mainnet is `1`.

Run the read-only configuration check as the installation owner:

```bash
az-vote-delegate check --config /absolute/path/config.json
```

Replace the path with your app configuration, and use `sudo` for a root installation. This checks the key, pillar, and chain ID without signing or publishing.

## A proposal or phase is missing from manual voting

The lists show only eligible, unanswered requests. A recorded yes, no, or abstain removes the item. Queued manual choices appear in **Pending transactions & manual overrides** instead.

A closed proposal is no longer eligible. A phase must be the current voting phase of an active project. Updated phases can have a new request ID. Refresh the menu to fetch current data.

## A manual vote says queued or pending

Queued is not the same as confirmed. If a copy vote was already submitted, the service waits for its confirmed contract receipt before applying your choice.

An ambiguous publish result remains pending to avoid duplicate votes. Inspect the transaction hash on your trusted node. Do not delete state or repeatedly submit through another tool to force progress. If the request closes first, a queued vote cannot reopen it.

## The service stops when I log out

For a user service, enable lingering for the node's account. See [regular user service setup](./service.md#regular-user-installation). Also confirm you installed a service instead of running the worker only in a foreground terminal.

## The menu cannot find my setup or service

Open it as the same user and with the same configuration used during setup. Root and regular user installations have different home directories. An existing `config.json` in your current directory can also select a different configuration; pass `--config` to remove ambiguity.

## I installed an update but the service uses the old version

The curl installer replaces the file, but an already running service needs a restart. Follow the matching root or user command in [Updates & upgrades](./updates.md#upgrade-with-curl).

If an in-app upgrade reports a restart failure, the new executable may already be installed. Check the service logs and restart the existing unit. Custom units and foreground workers need a manual restart.

## I cannot leave a menu

Press **Esc** or select **Back** to return one screen. At the main menu, Esc exits. **Ctrl+C** exits from any screen. The background service keeps running.

## I need more help

Check the [command reference](./commands.md) and [GitHub issues](https://github.com/0x3639/az-vote-delegate/issues). When reporting a problem, include the app version, selected mode, and relevant error text. Remove passwords, key material, and other private details from anything you share.
