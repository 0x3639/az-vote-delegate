---
title: Manual votes & phases
description: Cast your own vote on an AZ proposal or phase payout while copying other requests.
---

A manual vote takes priority **only for that proposal or phase-payout request**. The service keeps running and copying other eligible requests. There is no pause for a number of days.

## Cast a vote

1. Open the terminal menu as the installation owner.
2. Choose **Vote on unanswered AZ proposals** or **Vote on unanswered phase payouts**.
3. Select a request and review its details. A phase includes its parent project and requested ZNN/QSR.
4. Choose **yes**, **no**, or **abstain**, then confirm.

The lists show only requests your pillar has not voted on. Any recorded yes, no, or abstain hides that request. Queued manual choices also leave the list and appear under **Pending transactions & manual overrides**. The app checks again before accepting your selection in case your pillar voted while the list was open.

In **Dry run**, the service does not submit votes, and the menu asks you to switch to **Live** before accepting a manual vote. The menu distinguishes a queued choice from a confirmed manual vote; a confirmation screen alone is not proof of on-chain execution.

Press **Esc** or choose **Back** to return without submitting an unconfirmed choice. In a searchable list, **Tab** applies a filter and **Ctrl+R** clears it.

## Your choice persists

The app saves an explicit manual override before acknowledging it. If you choose yes and the delegate later votes no, your manual yes stays protected.

Overrides are saved by **pillar name + request ID** in the local state file. They survive service restarts, reboots, and upgrades. Preserve that file when moving the installation. The background service is the only signer and state writer; the menu sends your choices to it.

An override stays in effect until you select **Resume copying** in **Pending transactions & manual overrides**, the request closes, or its confirmed votes reset. An updated phase with a new ID is a new request; an old phase's override does not carry over. Project and phase overrides are independent.

## If a copy vote is already in progress

A manual choice interrupts copy-vote preparation. If the service has already submitted a copy vote, it waits for the confirmed contract receipt before applying your manual choice.

An uncertain publish result stays pending instead of being blindly sent again. Inspect pending transaction hashes on your trusted node if necessary. A queued choice cannot reopen a closed request or undo a contract decision.

## Votes from another application

Use this app's menu when you want the service to know that a vote is manual. A vote cast elsewhere can have exactly the same value as the service's last vote, so the value alone cannot identify your intent. Another application can also submit a vote at the same time as this service.

For guaranteed manual control through an external tool, stop this service, resolve pending submissions, and leave it stopped while voting externally.
