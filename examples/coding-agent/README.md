# coding-agent: turn an API key into a tool-using coding agent

`coding_agent.py` is a ~110-line agent: one `bash` tool, a turn loop, prompt caching, and hard caps on turns and
dollars. The model drives a shell in a working directory until the task is done. It bills the Console API key in
`~/.config/arena/key` (set it up with [`docs/setup-keys.md`](../../docs/setup-keys.md)); it never touches a
Claude.ai or Claude Code subscription. Full writeup, cost numbers and lessons: [`docs/coding-agent.md`](../../docs/coding-agent.md).

## Run the demo

```bash
# Work in a throwaway directory — the agent runs real shell commands there with your permissions.
mkdir -p /tmp/agent-demo && : > /tmp/agent-demo/.keep
python3 -I coding_agent.py \
  --task demo-task.md --cwd /tmp/agent-demo --out ./out-demo \
  --model claude-opus-4-8 --max-turns 20 --budget-usd 2
cat out-demo/agent-cost.txt
```

You'll see per-turn cost tick up, the agent write and run a test, and a `FINAL:` line. The transcript (every command
and its output) is in `out-demo/agent-run.log`.

## Use it for real work

Write a `task.md` that says what to build, **where to write deliverables** (the `--out` directory), and **how to
verify** (exact test commands and expected results). Point `--cwd` at a disposable copy of the repo. Set a
`--budget-usd` you're willing to spend; the loop stops there and keeps whatever it finished.

```bash
python3 -I coding_agent.py --task task.md --cwd ./work-copy --out ./out --budget-usd 15
```

## The two rules that matter

- **Save as you go.** Tell the agent in `task.md` to write each finished piece to `--out` *before* starting the next.
  If the budget or turn cap hits mid-way, unsaved work is only in the working copy, not packaged. (We learned this the
  expensive way — see the doc.)
- **Disposable `--cwd`.** There is no sandbox. The bash tool is real. Give it a copy you can throw away.
