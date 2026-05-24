# Agent Memory Protocol

To ensure continuity across multiple agents and developers, all agents MUST maintain the `MEMORY.md` file at the root of the repository.

## Workflow
1. **Read First**: Before starting any task, read `MEMORY.md` to understand the current state of the project, recent decisions, and known hurdles.
2. **Update Always**: Update `MEMORY.md` immediately after completing a significant logical step or when a key architectural decision is made.
3. **Hand-off**: Before finishing a session, summarize the current state and leave "clues" for the next agent.

## Memory Format
The `MEMORY.md` file must follow this structure:

### 🟢 Current Focus
- **Objective**: [Short description of the current goal]
- **Active Issue**: [#IssueNumber]
- **Status**: [In-progress / Blocked / Verifying]

### 🏗️ Architecture & Decisions
- **Key Decisions**: [Decision] -> [Reason/link to ADR]
- **Current State**: [Brief summary of what is currently implemented and working]

### ⚠️ Hurdles & Gotchas
- **Blockers**: [What is stopping progress]
- **Technical Debt**: [Known shortcuts taken that need fixing]
- **Edge Cases**: [Discovered behaviors that need handling]

### 💡 Working Hypotheses
- [Assumption being tested] -> [Expected result]

### 📡 Hand-off Notes
- **Next Steps**: [Actionable next step for the next agent]
- **Warning**: [Specific files or functions to be careful with]
