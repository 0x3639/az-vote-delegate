First release of the Accelerator-Z vote manager for Zenon pillar operators.

- One-command installer with SHA-256 verification; static Linux amd64 and arm64 binaries.
- NOMCTL-style terminal menu and setup from an existing znnd producer configuration.
- Searchable delegate selection and persistent background service.
- Copy voting and manual yes/no/abstain for project proposals and phase-payout requests.
- Explicit manual overrides, including same-value votes, coordinated by one signer.
- Dry-run by default, persistent pending-transaction tracking and guarded publication.

Manual choices wait for previously submitted votes to execute. They can be applied only while the request remains open. Votes submitted by external tools still have the off-chain detection limitations described in the README.

No live producer-key or mainnet acceptance run was performed. Automated checks use native wallet-format conformance tests and local RPC fixtures.
