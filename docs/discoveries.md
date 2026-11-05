# Discoveries

What we learned building and running the chess arena (November 2026). API facts change; check the official docs before relying on any of these.

## Prices (USD per million tokens)

| Model | Input | Output | Cache read | Cache write (5 min) |
|---|---|---|---|---|
| `claude-haiku-5-5` | 0.10 | 0.50 | 0.01 | 0.125 |
| `claude-sonnet-5-5` | 2.00 | 10.00 | 0.10 | 2.50 |
| `claude-opus-5-5` | 4.00 | 20.00 | 0.20 | 5.00 |
| `claude-fable-5-1` | 10.00 | 50.00 | 0.25 | 12.50 |

- A 5-minute cache write costs 1.25× input; a 1-hour cache write costs 2×.
- Haiku 5.5 has a price step: a request with more than 100K prompt tokens costs 5× ($0.50 / $2.50). Keep prompts small and stateless.
- Thinking tokens bill as output tokens.
- Keep prices in a config file, not in code.

## Request rules that will bite you

- **Forced tool choice is rejected.** `tool_choice` of `any` or `tool` returns a 400 on Opus 5.5, Sonnet 5.5 and Fable 5.1. Use `auto`, mark the tool `strict: true`, set `disable_parallel_tool_use: true`, and tell the model in the system prompt to call the tool exactly once.
- **Thinking can't be turned off on Opus 5.5.** `{type: "disabled"}` is a 400 at every effort level; lower `effort` instead. Sonnet 5.5 also rejects disabled thinking. Simplest rule: never send `thinking: disabled`.
- **Effort defaults to `medium` on Opus 5.5 and Haiku 5.5.** Pin it in config if you compare models.
- **Don't send `temperature`, `top_p` or `top_k`.** Several current models reject non-default values.
- **Server-side fallbacks are opt-in** (a `fallbacks` parameter plus a beta header). If you measure per model, never enable them, and check that `response.model` equals the model you asked for. In the arena a mismatch stops the run.
- **The SDK retries for you.** The Go SDK retries 408/409/429/5xx twice by default. Turn that off (`WithMaxRetries(0)`) if your own code counts retries, or your counts will be wrong.

## Prompt caching

- Put stable content first: the tool schema, then the system prompt, with the cache breakpoint at the end of the system prompt. Put everything that changes per call into the one user message.
- Never put per-turn data in the tool schema (for example legal moves as an `enum`); it breaks the cache on every call.
- The minimum cacheable prefix on the newest models is 512 tokens, so a ~1.5K-token system prompt caches.
- Measured: about 1.5K tokens read from cache on every chess decision after the first, with zero cache misses across 247 decisions.

## Measured: Haiku 5.5 playing chess vs a random-legal bot

Two runs, 247 decisions, effort `medium`, 6 finished games:

| Metric | Value |
|---|---|
| Results | 5 wins by checkmate (43–107 plies), 1 draw by stalemate (Haiku was winning) |
| Cost per move | $0.0008 |
| Cost per game | $0.015–0.05 |
| Latency per move | 8.5 s mean, about 15 s p95 |
| Tokens per decision | ~800 fresh input, ~1.5K cached input, ~1.4K output (mostly thinking) |
| Illegal attempts | 2 of 249 (0.8%); each fixed on the retry |
| Timeouts / provider errors | 0 / 0 |

Takeaways:

- Haiku is cheap enough that the whole Haiku-vs-random gate costs well under $1.
- Games are slow, not expensive: about 5–10 minutes each at 2 games in parallel.
- Most output tokens are thinking. Effort is the main cost lever.
- Planning estimates of "~40 decisions per game" are too low for weak play; games often run 50–100+ plies.

## Harness gotchas

- **Go `signal.NotifyContext` cancels with a cause.** If you check `context.Cause(ctx)` to detect fatal errors, Ctrl-C looks like one. Decide from whether the parent context was cancelled.
- **`go run` binaries live in `~/.cache/go-build/...`**, not `/tmp/go-build`, so a `pgrep` for the temp path misses them. Find the child of the `go run` process instead.
- **`pgrep -f <pattern>` inside a shell command can match that shell itself.** Pick the right PID before sending signals.
- **Poll the game server; don't trust your local copy.** Keep a local board only for display (SAN history) and check it against the server's position after every move.
