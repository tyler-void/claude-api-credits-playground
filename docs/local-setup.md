# Local setup, step by step

From zero to one paid API call on your machine. Takes about 10 minutes on macOS, Linux or Windows.

Commands come in two flavours:

- **macOS / Linux:** any POSIX shell (zsh, bash). On Windows with WSL, follow these too, inside WSL.
- **Windows:** PowerShell (Windows PowerShell 5.1 or PowerShell 7). `cmd.exe` is not covered.

## 1. Install the tools

You need **Go 1.25 or newer** and **git**.

| OS | Install |
|---|---|
| macOS | `brew install go git` (or the installer from [go.dev/dl](https://go.dev/dl/)) |
| Linux | Your package manager, if it ships Go 1.25+; otherwise the tarball from [go.dev/dl](https://go.dev/dl/) |
| Windows | `winget install GoLang.Go Git.Git`, then open a new PowerShell window |

Check: `go version` should print 1.25 or newer.

## 2. Get the code

```bash
git clone https://github.com/tyler-void/claude-api-credits-playground.git
cd claude-api-credits-playground
```

Same commands in PowerShell.

## 3. Create the key in the Console

1. Sign in to the Claude Console with the organization that holds your plan's API credit.
2. Create a workspace for experiments, for example `arena`.
3. In that workspace's settings, set a **monthly spend limit**. It's the backstop if a script runs away.
4. Create an API key **in that workspace** and copy it. The Console shows it only once.

If you have more than one key (say a second $200 one), give each its own workspace and limit. See [Two or more keys](#two-or-more-keys).

## 4. Save the key in a private file

Every OS uses the same location: `~/.config/arena/key` (on Windows, `$HOME\.config\arena\key`). Make the file private **before** you paste the key into it.

**macOS / Linux**

```bash
mkdir -p ~/.config/arena
install -m 600 /dev/null ~/.config/arena/key   # owner-only from the start
${EDITOR:-nano} ~/.config/arena/key            # paste the key, save, quit
```

Check it:

```bash
stat -c '%a' ~/.config/arena/key      # Linux, expect 600
stat -f '%Lp' ~/.config/arena/key     # macOS, expect 600
```

**Windows (PowerShell)**

```powershell
$f = "$HOME\.config\arena\key"
New-Item -ItemType File -Force $f | Out-Null
icacls $f /inheritance:r /grant:r "${env:USERNAME}:(F)" | Out-Null   # only you can read it
notepad $f                                                          # paste the key, save, close
```

Check it:

```powershell
icacls "$HOME\.config\arena\key"
# expect one line granting your user (F), and nothing for Users, Everyone or Administrators
```

Don't:

- put the key in `~/.zshrc`, `~/.bashrc`, your PowerShell `$PROFILE`, Windows "Environment Variables" settings, or a `.env` file inside a repo;
- type it on a command line or paste it into an AI chat;
- copy it into `ANTHROPIC_API_KEY`, which other tools on your machine also read.

## 5. Run the hello example

**macOS / Linux**

```bash
cd examples/hello-credits
ARENA_ANTHROPIC_API_KEY="$(cat ~/.config/arena/key)" go run . "Say hi in five words."
```

`$(cat ...)` hands the key to this one process only. It doesn't stay in your shell, and the command line saved in history doesn't contain it.

**Windows (PowerShell)**

PowerShell has no `VAR=value command` form, so set the variable, run, and remove it again even if the run fails:

```powershell
cd examples\hello-credits
try {
  $env:ARENA_ANTHROPIC_API_KEY = (Get-Content -Raw "$HOME\.config\arena\key").Trim()
  go run . "Say hi in five words."
} finally {
  Remove-Item Env:ARENA_ANTHROPIC_API_KEY -ErrorAction SilentlyContinue
}
```

`.Trim()` drops the trailing newline Notepad may save. History records the command, not the key.

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

## Two or more keys

Keep one file per key, next to each other, each made exactly as in step 4:

```
~/.config/arena/key     # first key,  workspace "arena"
~/.config/arena/key2    # second key, its own workspace and spend limit
```

Pick the key per run by changing only the file name:

```bash
ARENA_ANTHROPIC_API_KEY="$(cat ~/.config/arena/key2)" go run . "Say hi in five words."
```

```powershell
$env:ARENA_ANTHROPIC_API_KEY = (Get-Content -Raw "$HOME\.config\arena\key2").Trim()
```

The Console usage page shows which workspace, and so which key, paid for each run.

## Troubleshooting

| Error | Likely cause |
|---|---|
| `ARENA_ANTHROPIC_API_KEY is not set` | The key file is empty or the path is wrong |
| API returned 401 | Key typo, extra characters in the file (on Windows, did you `.Trim()`?), or the key was revoked |
| API returned 403 | The key's workspace or organization can't use that model |
| API returned 429 | Rate limit or spend limit reached; wait, or raise the workspace limit |
| API returned 400 | A request the model rejects (see `docs/discoveries.md`, "Request rules") |
| `icacls` says access denied | Run it as the same user who created the file, not as Administrator |
| `go: command not found` / not recognized | Open a new terminal after installing Go, so `PATH` picks it up |

## Rotate or remove a key

1. Create a new key in the Console and save it over the file (step 4).
2. Delete the old key in the Console.
3. When you're done for good, delete the file: `rm ~/.config/arena/key` (PowerShell: `Remove-Item "$HOME\.config\arena\key"`).
