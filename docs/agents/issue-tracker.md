# Issue tracker: GitHub

Issues and PRDs for this repo live as GitHub issues. Use the `gh` CLI for all operations.

## Conventions

Every issue must satisfy all three requirements before it is considered ready:

1. **Triage label** — Exactly one of the five canonical labels (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.
2. **Title prefix** — Signal the work type via the title prefix convention (`PRD:`, `Spike:`, `Docs:`, `Bug:`, or no prefix). See `docs/project-board.md`.
3. **Size and decomposition** — Set `Size` per the definitions in `docs/project-board.md`. If the size is L or XL (or M that is not a single coherent vertical slice), **do not mark it `ready-for-agent`**. Decompose it into XS/S sub-issues first using the [Decomposition Policy](docs/project-board.md#decomposition-policy) before applying `ready-for-agent` to individual sub-issues.
4. **Project fields** — When added to the project board, set `Status`, `Priority`, `Area`, and `Size`. Keep `Status` accurate through the lifecycle. Field definitions are in `docs/project-board.md`.

- **Create an issue**: `gh issue create --title "..." --body "..."`. Use a heredoc for multi-line bodies.
- **Read an issue**: `gh issue view <number> --comments`, filtering comments by `jq` and also fetching labels.
- **List issues**: `gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'` with appropriate `--label` and `--state` filters.
- **Comment on an issue**: `gh issue comment <number> --body "..."`
- **Apply / remove labels**: `gh issue edit <number> --add-label "..."` / `--remove-label "..."`
- **Close**: `gh issue close <number> --comment "..."`

Infer the repo from `git remote -v` - `gh` does this automatically when run inside a clone.

## When a skill says "publish to the issue tracker"

Create a GitHub issue in `supperjumpin/supperjumpin`.

Add implementation issues and PRDs to the Supperjumpin GitHub Project when practical: https://github.com/orgs/supperjumpin/projects/1. Project field conventions are documented in `docs/project-board.md`.

When breaking down a PRD, prefer GitHub sub-issues under the PRD issue for implementation slices that directly deliver that PRD. If the PRD is scoped to a milestone, assign that milestone to the PRD issue and each sub-issue explicitly — see `docs/project-board.md` for milestone conventions. Use normal first-order issues for cross-cutting tech decisions, infrastructure chores, bugs, or future ideas that may support multiple PRDs.

When creating issues on behalf of the current operator, assign them to the current GitHub user. `gh issue edit <numbers...> --add-assignee @me` is the default unless the user says otherwise.

When creating sub-issues, set both relationship surfaces before calling the work done:

- **Parent relationship**: attach each implementation issue as a GitHub sub-issue of the parent PRD using GraphQL `addSubIssue(input: { issueId: <parent_issue_node_id>, subIssueId: <child_issue_node_id>, replaceParent: true })`.
- **Blocker relationship**: for implementation dependencies, add GitHub's relationship field through GraphQL `addBlockedBy(input: { issueId: <blocked_issue_node_id>, blockingIssueId: <blocker_issue_node_id> })`.

Fetch node IDs with GraphQL before applying relationships:

```sh
gh api graphql -f query='query { repository(owner:"supperjumpin", name:"supperjumpin") { parent: issue(number: 268) { id number title } child: issue(number: 269) { id number title } } }'
```

Then apply the relationships:

```sh
gh api graphql \
  -f query='mutation($parent: ID!, $sub: ID!) { addSubIssue(input: {issueId: $parent, subIssueId: $sub, replaceParent: true}) { issue { number } subIssue { number } } }' \
  -F parent='<parent_issue_node_id>' \
  -F sub='<child_issue_node_id>'

gh api graphql \
  -f query='mutation($issue: ID!, $blocking: ID!) { addBlockedBy(input: {issueId: $issue, blockingIssueId: $blocking}) { issue { number } blockingIssue { number } } }' \
  -F issue='<blocked_issue_node_id>' \
  -F blocking='<blocker_issue_node_id>'
```

Verify the graph after creating it:

```sh
gh api graphql -f query='query { repository(owner:"supperjumpin", name:"supperjumpin") { issue(number: 268) { number title subIssues(first: 20) { nodes { number title } } } } }'
```

Keep the `## Parent` and `## Blocked by` sections in the issue body too. The structured GitHub relationships drive project views; the body keeps the dependency readable in CLI output and notifications.

## When a skill says "fetch the relevant ticket"

Run `gh issue view <number> --comments`.
