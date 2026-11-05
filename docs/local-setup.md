# Local setup, step by step

From zero to one paid API call on your machine. Takes about 10 minutes. Linux and macOS.

## 1. Install the tools

- **Go 1.25 or newer:** `go version` should print 1.25+.
- **git**

## 2. Get the code

```bash
git clone https://github.com/tyler-void/claude-api-credits-playground.git
cd claude-api-credits-playground
```

## 3. Create the key in the Console

1. Sign in to the Claude Console with the organization that holds your plan's API credit.
2. Create a workspace for experiments, for example `arena`.
3. In that workspace's settings, set a **monthly spend limit**. It's the backstop if a script runs away.
4. Create an API key **in that workspace** and copy it. The Console shows it only once.

If you have more than one key (say a second $200 one), give each its own workspace and limit.

## 4. Save the key in a private file

```bash
mkdir -p ~/.config/arena
install -m 600 /dev/null ~/.config/arena/key   # owner-only from the start
$EDITOR ~/.config/arena/key                     # paste the key, save, quit
```

Check it:

```bash
stat -c '%a' ~/.config/arena/key     # macOS: stat -f '%Lp' ~/.config/arena/key
# expect 600
```

For a second key, use a second file, for example `~/.config/arena/key-2`, made the same way.

Don't:

- put the key in `~/.zshrc`, `~/.bashrc` or a `.env` file inside a repo;
- type it on a command line or paste it into an AI chat;
- copy it into `ANTHROPIC_API_KEY`, which other tools on your machine also read.

## 5. Run the hello example

```bash
cd examples/hello-credits
ARENA_ANTHROPIC_API_KEY="$(cat ~/.config/arena/key)" go run . "Say hi in five words."
```

`$(cat ...)` hands the key to this one process only. It doesn't stay in your shell, and the command line saved in history doesn't contain it.

Expected output, costing a fraction of a cent on Haiku:

```
Hi there, nice to meet you!

model claude-haiku-5-5  stop end_turn  latency 1.9s
tokens: input 14, output 12, cache read 0, cache write 0
estimated cost: $0.000007
```

Other models: add `-model claude-opus-5-5 -effort low`, for example. The prices come from `prices.yaml`.

## 6. Check it worked

- In the Console, the workspace's usage page shows the call within a few minutes.
- If the reply came from a different model than you asked for, the example prints a warning.

## Troubleshooting

| Error | Likely cause |
|---|---|
| `ARENA_ANTHROPIC_API_KEY is not set` | The key file is empty or the path is wrong |
| API returned 401 | Key typo, extra characters in the file, or the key was revoked |
| API returned 403 | The key's workspace or organization can't use that model |
| API returned 429 | Rate limit or spend limit reached; wait, or raise the workspace limit |
| API returned 400 | A request the model rejects (see `docs/discoveries.md`, "Request rules") |

## Rotate or remove a key

1. Create a new key in the Console and save it over the file (step 4).
2. Delete the old key in the Console.
3. When you're done for good: `rm ~/.config/arena/key`.
