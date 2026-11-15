# Using an API key as a coding agent

The experiments in `PLAN.md` are games. This note covers a different, very practical use of the same credit: driving
a **tool-using coding agent** with a Console API key, so heavy agentic work (reading a repo, editing, running tests)
bills your API credit instead of a Claude.ai / Claude Code subscription. The whole thing is one file,
[`examples/coding-agent/coding_agent.py`](../examples/coding-agent/coding_agent.py) — no SDK, no framework.

It was built and run for real: solving a batch of graded SDLC coding tasks to produce a reference dataset. The cost
and failure numbers below are measured from that run (opus-4-8), not estimated.

## How it works

A coding agent is just a loop around the Messages API with **tool use**:

1. You declare one tool, `bash`, with a JSON schema (`{cmd: string}`).
2. You send the system prompt + the task brief + the conversation so far, with `tools: [bash]`.
3. The model replies either with text, or with a `tool_use` block asking to run a command.
4. You run that command in a fixed working directory and send the combined stdout+stderr back as a `tool_result`.
5. Repeat until the model emits a line starting `FINAL:` (our sentinel), or you hit a turn / dollar cap.

One `bash` tool is enough for everything — `cat`/`sed` to read and edit, `python3`/`npx vitest` to run, `git diff`
to inspect. You don't need separate read/edit/write tools; the shell is the tool.

Key points in the implementation:

- **Billing.** Every call uses `x-api-key` from a key file. That is the Console key, so it draws your API credit.
  Nothing here reads a subscription token. (Confirm which you're spending: the API-key path never shows a "Claude
  Code" or OAuth login; a 401 `invalid x-api-key` means the key file is wrong, not that it fell back to a subscription.)
- **Prompt caching.** The task brief is sent every turn and never changes, so it carries
  `cache_control: {type: ephemeral}` and the beta header `anthropic-beta: prompt-caching-2024-07-31`. After turn 1 the
  brief is billed at the cache-**read** rate (10× cheaper than input), which matters a lot once the transcript is long.
- **Caps.** `--max-turns` and `--budget-usd` both stop the loop. The running cost is computed from each response's
  `usage` block against a small price table and printed per turn.
- **Transcript.** Every command and its output is appended to `<out>/agent-run.log`, so a finished or stopped run is
  fully auditable.

## What it cost (measured, opus-4-8, Nov 2026)

Pricing used (USD per 1M tokens): input 5, cache-write 6.25, cache-read 0.5, output 25.

| Run | Turns | Cost | Notes |
|---|---|---|---|
| smoke test (2 trivial commands) | 2 | $0.01 | proves the loop |
| one `bash` turn, early (small context) | — | ~$0.01 | |
| heavy group of 2–3 coding items, with many test runs | 58–89 | **$16–$35** | context grows each turn; this dominates |
| four such groups + four read-only "audit" calls | — | **≈ $101 total** | |

Takeaways:

- **Turns, not calls, drive cost.** A single call is cents; an 85-turn agentic session on a large repo is tens of
  dollars because the whole transcript is re-sent every turn. Prompt caching softens the fixed brief, but the growing
  tool-result history is billed at input rate. Budget per *session*, and cap it.
- **opus is premium.** For cheaper runs, switch `--model` to `claude-sonnet-4-5` (output 15 vs 25, input 3 vs 5) —
  often fine for mechanical coding, keep opus for the hard reasoning.
- **A read-only "second opinion" call is cheap.** Feeding a brief + the spec + the current code and asking for a
  rule-by-rule critique (no tools, one call) cost ~$0.15 each and was a useful cross-check on the agent's work.

## Lessons (the expensive ones)

- **Save deliverables as you go.** One group spent its full $35 budget and stopped mid-verification *after* finishing
  the work but *before* copying it to the output directory — so the output folder was empty and only the working copy
  held the result. Always instruct the agent (in the brief) to write each finished item to `--out` before starting the
  next, and set the cap high enough to include packaging. Recovering from the working copy is possible but annoying.
- **Credit exhaustion looks like an HTTP 400.** When the balance runs out mid-run the API returns
  `400 invalid_request_error: "Your credit balance is too low to access the Anthropic API."` The loop exits on it; any
  completed work is still on disk. Check balance before a big batch.
- **A dead key fails fast with 401.** `401 invalid x-api-key` — a revoked/expired key, not a transient error; don't
  retry it. (We had two keys; one was dead from the start, so everything ran on the other.)
- **Model-specific request quirks.** On opus-4-8, `thinking: {type: "enabled", budget_tokens: N}` is rejected
  (`use "thinking.type.adaptive" and "output_config.effort"`). The simplest robust path is to send no `thinking`
  block at all; the agent still reasons fine across turns.
- **Run two sessions at once, not ten.** Parallel sessions share the machine's CPU (each spawns real test runs) and
  the account's rate limits. Two concurrent heavy sessions was a good balance; four thrashed.

## When to reach for this vs. the subagent built into your tool

If you're inside Claude Code, its own subagents are simpler and need no key — but they bill your *subscription usage*.
When you have API credit to spend and want to keep subscription usage down (or want many parallel workers, or a
second model's opinion), this script turns that credit into exactly the same kind of tool-using worker. Same model,
different wallet.
