#!/usr/bin/env python3
"""A minimal tool-using coding agent on the Claude Messages API, billed to a Console API key.

No SDK, no framework: one file, one tool (bash), a turn loop, prompt caching, and hard caps on turns and dollars.
This is the whole trick behind "make my API key act like a coding agent" — the model asks to run shell commands,
you run them in a working directory and feed the output back, and it edits/builds/tests through the shell until it
says FINAL. Everything bills the key in ~/.config/arena/key, never a Claude.ai / Claude Code subscription.

    python3 -I coding_agent.py \
        --task task.md --cwd /path/to/repo --out ./out \
        [--key ~/.config/arena/key] [--model claude-opus-4-8] [--max-turns 60] [--budget-usd 10]

task.md is a plain-language brief: what to build, where to write it, how to verify. The agent's cwd for every shell
command is --cwd. It stops on a line beginning "FINAL:", at --max-turns, or when the running cost reaches --budget-usd.
Per-turn cost and a full transcript are written to <out>/agent-run.log.

Safety: the bash tool runs real commands with your permissions in --cwd. Point it at a disposable working copy, not
your only copy. There is no sandbox here; a timeout per command is the only guard.
"""
import argparse, json, os, subprocess, sys, time, urllib.request, urllib.error

ap = argparse.ArgumentParser()
ap.add_argument('--task', required=True, help='brief file (plain text/markdown)')
ap.add_argument('--cwd', required=True, help='working directory for every shell command')
ap.add_argument('--out', required=True, help='where the agent writes deliverables + the run log')
ap.add_argument('--key', default='~/.config/arena/key', help='file holding the API key')
ap.add_argument('--model', default='claude-opus-4-8')
ap.add_argument('--max-turns', type=int, default=60)
ap.add_argument('--budget-usd', type=float, default=10.0)
ap.add_argument('--cmd-timeout', type=int, default=300, help='seconds per shell command')
a = ap.parse_args()

CWD = os.path.abspath(os.path.expanduser(a.cwd))
OUT = os.path.abspath(os.path.expanduser(a.out))
os.makedirs(OUT, exist_ok=True)
KEY = open(os.path.expanduser(a.key)).read().strip()
# List price per 1M tokens: input, cache-write, cache-read, output (USD). Update from docs/discoveries.md.
PRICE = {'claude-opus-4-8': (5, 6.25, 0.5, 25),
         'claude-sonnet-4-5': (3, 3.75, 0.3, 15)}.get(a.model, (5, 6.25, 0.5, 25))

BASH = {'name': 'bash',
        'description': 'Run a bash command in the working directory and get back combined stdout+stderr.',
        'input_schema': {'type': 'object', 'properties': {'cmd': {'type': 'string'}}, 'required': ['cmd']}}

# The brief is marked cache_control so re-sending it every turn is cheap (cache-read price, not input price).
system = [
    {'type': 'text', 'text':
        'You are an expert engineer working through one task. You have a single tool, bash, whose working directory '
        f'is fixed. Explore, edit, build and test with it (cat, sed, python3, npm/npx, git diff, ...). Verify your '
        f'work with real runs. Write deliverables under {OUT}. When the task is fully done, reply with a line that '
        'starts "FINAL:" and a short summary. Be economical: read only what you need, cap long output with | tail, '
        'and never re-read a file you already have.'},
    {'type': 'text', 'text': open(os.path.expanduser(a.task), encoding='utf8').read(),
     'cache_control': {'type': 'ephemeral'}},
]


def bash(cmd):
    try:
        p = subprocess.run(['bash', '-lc', cmd], cwd=CWD, capture_output=True, text=True, timeout=a.cmd_timeout)
        out = p.stdout + p.stderr
    except subprocess.TimeoutExpired:
        out = f'[timed out after {a.cmd_timeout}s]'
    except Exception as e:
        out = f'[error: {e}]'
    return out if len(out) <= 24000 else out[:12000] + '\n...[clipped]...\n' + out[-8000:]


def call(messages):
    body = json.dumps({'model': a.model, 'max_tokens': 8000, 'system': system,
                       'tools': [BASH], 'messages': messages}).encode()
    req = urllib.request.Request('https://api.anthropic.com/v1/messages', data=body, headers={
        'x-api-key': KEY, 'anthropic-version': '2023-06-01', 'content-type': 'application/json',
        'anthropic-beta': 'prompt-caching-2024-07-31'})
    for attempt in range(4):
        try:
            return json.load(urllib.request.urlopen(req, timeout=600))
        except urllib.error.HTTPError as e:
            msg = e.read().decode()[:200]
            if e.code in (429, 500, 529) and attempt < 3:
                time.sleep(5 * (attempt + 1)); continue
            sys.exit(f'HTTP {e.code}: {msg}')     # e.g. 400 "credit balance is too low", 401 invalid key
        except Exception as e:
            if attempt < 3:
                time.sleep(5); continue
            sys.exit(str(e))


def cost(u):
    return (u.get('input_tokens', 0) * PRICE[0] + u.get('cache_creation_input_tokens', 0) * PRICE[1]
            + u.get('cache_read_input_tokens', 0) * PRICE[2] + u.get('output_tokens', 0) * PRICE[3]) / 1e6


messages = [{'role': 'user', 'content': 'Begin. Read the brief, then work. Reply FINAL: when fully done.'}]
spent = 0.0
log = open(f'{OUT}/agent-run.log', 'w')
for turn in range(1, a.max_turns + 1):
    r = call(messages)
    spent += cost(r.get('usage', {}))
    blocks = r.get('content', [])
    messages.append({'role': 'assistant', 'content': blocks})
    said = ''.join(b.get('text', '') for b in blocks if b.get('type') == 'text')
    tools = [b for b in blocks if b.get('type') == 'tool_use']
    line = f'turn {turn} ${spent:.2f} stop={r.get("stop_reason")} tools={len(tools)}'
    if said.strip():
        line += ' | ' + said[:160].replace('\n', ' ')
    print(line, flush=True); log.write(line + '\n'); log.flush()
    if 'FINAL:' in said or (r.get('stop_reason') == 'end_turn' and not tools):
        print(f'DONE in {turn} turns, ${spent:.2f}'); break
    if spent >= a.budget_usd:
        print(f'BUDGET ${a.budget_usd} reached at turn {turn}; stopping. Deliverables so far are in {OUT}.'); break
    results = []
    for t in tools:
        out = bash(t['input'].get('cmd', ''))
        log.write(f'$ {t["input"].get("cmd", "")}\n{out}\n'); log.flush()
        results.append({'type': 'tool_result', 'tool_use_id': t['id'], 'content': out})
    messages.append({'role': 'user', 'content': results or 'Continue, or reply FINAL: when done.'})
else:
    print(f'hit turn cap {a.max_turns}, ${spent:.2f}')
open(f'{OUT}/agent-cost.txt', 'w').write(f'{a.model} ${spent:.2f}\n')
