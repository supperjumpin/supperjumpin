# Coverage as a Visible Signal

Coverage in Supperjumpin is a reporting and review signal, not a merge gate. For the Go API and `api-client`, CI should surface coverage in the GitHub Actions summary and leave a PR comment when coverage drops from the previous baseline, so the signal is visible without slowing people down.

This keeps coverage actionable while avoiding threshold churn before launch. Mobile is still excluded from the coverage policy for now, but its baseline has moved from manual verification to a lightweight Jest plus React Native Testing Library harness alongside typecheck.
