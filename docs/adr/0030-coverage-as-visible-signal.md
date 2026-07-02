# Coverage as a Visible Signal

Coverage in Supperjumpin is a reporting and review signal, not a merge gate. For the Go API and the Discord bot, CI should surface coverage in the GitHub Actions summary and leave a PR comment when coverage drops from the previous baseline, so the signal is visible without slowing people down. Coverage is folded into `mage test -coverage` (ADR-0047), so a single invocation produces the standard `go test` output plus a per-package summary and a total percent.

This keeps coverage actionable while avoiding threshold churn before launch. With the mobile app and api-client gone (ADR-0048, ADR-0049), the only coverage surfaces are the two Go services.
