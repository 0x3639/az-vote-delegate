Makes repeat installations explicit and verifies persistent manual-vote isolation.

- The installer detects an existing executable and replaces it at the same path. It preserves configuration, voting mode and saved manual overrides, and does not start an additional service.
- Upgrade output prints the restart command for the existing system or user service. The running service uses the old executable until restarted.
- Documents that normal upgrades use the same user, install directory and configuration.
- Adds regression coverage for repeated installation and for manual AZ/phase votes surviving disk reloads while other proposals and phases continue copying. No voting-engine behavior changes were needed.

Manual overrides are per pillar and request ID, with no timed service pause. They remain until copying is explicitly resumed, the request closes, or its confirmed votes reset. Different phase IDs have independent overrides.

**Compatibility:** This release supports the legacy fused-plasma / proof-of-work voting flow. Dynamic Plasma is not supported.
