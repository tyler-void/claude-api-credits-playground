# Notes for coding agents

- Read `PLAN.md` and `docs/discoveries.md` before changing anything.
- The experiments key is passed only as `ARENA_ANTHROPIC_API_KEY`, built from a private file: `ARENA_ANTHROPIC_API_KEY="$(cat ~/.config/arena/key)" ...`. Never read `ANTHROPIC_API_KEY`, never print, log or commit a key, never ask the user to paste one into chat.
- Build SDK clients with `option.WithoutEnvironmentDefaults()` plus an explicit `option.WithAPIKey`.
- Tests use fake HTTP servers; no real API calls in tests or CI.
- Real runs spend money: confirm with the user before starting one, and set `max_usd`.
- Only the official route: Console API keys. Do not add anything that extracts or reuses Claude.ai or Claude Code login tokens.
- Commit with the repo's git identity; don't add personal names or emails to files.
