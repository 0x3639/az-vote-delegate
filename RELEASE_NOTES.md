Makes voting mode an explicit choice during setup and in the main menu.

- Choose **Dry run — preview without submitting votes** or **Live — submit votes**. Dry run is selected by default for new setups.
- Setup then shows the chosen mode and a separate **Save and start** (or **Save setup**) / **Cancel** confirmation. Cancel never selects live mode.
- The service starts in the confirmed mode. The main menu's **Voting mode: Dry run / Live** option uses the same explicit choices.
- Explains password storage separately in plain language, without mixing file-permission details into the mode question.

Existing settings are preserved. Switching to dry run stops new submissions; transactions already submitted may still finish. Both modes apply to AZ proposals and phase payouts.

**Compatibility:** This release supports the legacy fused-plasma / proof-of-work voting flow. Dynamic Plasma is not supported.
