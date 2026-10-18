---
title: Development & releases
description: Build and test AZ Vote Delegate, and package verified Linux releases.
---

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

Pushing a version tag runs [the release workflow](https://github.com/0x3639/az-vote-delegate/blob/main/.github/workflows/release.yml):
tests and lint on Linux, static CGO builds for amd64 and arm64, smoke tests on
both (arm64 via QEMU), archive/checksum creation and a draft GitHub release.
Publish that draft only after the build and asset checks pass.

Live mainnet/testnet acceptance remains separate. Before operating with real
votes, run `check` and review dry-run output on the intended node. On an isolated
testnet, verify project yes→no copying, same-value manual takeover, a manual vote
while copying, phase-payout copying/manual override and a phase revision. No
live voting acceptance is implied by the offline tests.

The copy-delay tests use a controlled clock and mock votes to cover project and phase timing, restart recovery, changed or withdrawn source votes, round resets, closed requests, manual overrides, dry-run behavior, and live setting changes. They never wait real minutes or submit real votes.
