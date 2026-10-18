---
title: Choose a delegate
description: Select a pillar to copy using its voting participation.
---

Select a pillar during setup, or choose **Choose pillar to copy** from the main menu. The service copies that pillar's votes on eligible AZ proposals and phase-payout requests.

## The top 30 voters

The picker ranks active pillars by the number of distinct requests they have voted on, excludes your own pillar, and takes the top 30. If fewer are available, it shows them all.

The list is shuffled each time you open it. Ties are randomly ordered before the cutoff, so names near the start of the alphabet do not get priority.

An option may show:

```text
82.0% participation · 82/100 requests
```

Yes, no, and abstain all count as participation. The denominator includes all project proposals and retained phase requests returned by the node, including closed requests. It is not adjusted for when a pillar joined. Replaced or deleted phase revisions cannot be included when the node no longer returns them.

Participation tells you how often a pillar has voted in the available history. It does not tell you whether you agree with its decisions. With no recorded requests, the menu reports that instead of inventing a percentage.

## What gets copied

The service checks current votes on each poll. It follows open initial project proposals within their voting window and the current voting phase of active projects. If the delegate has not voted, there is no vote to copy yet.

Project proposals and phase payouts have separate request IDs. A vote on a project does not automatically count as a vote on its phase. When an updated phase receives a new votable ID, the app treats it as a fresh voting round.

The app votes on payout requests. It does not submit a project owner's payment request or guarantee an immediate payment; the contract rules and other pillars' votes still determine the result.

## Change who you follow

Choose **Choose pillar to copy**, select a new pillar, and confirm. Your saved [manual overrides](./manual-voting.md) stay in place. Other eligible requests follow the new delegate.

Delegation chains can work, but an empty cycle where two pillars only follow each other produces no initial vote. Choose a pillar that makes voting decisions.
