# Setting up a key that spends your plan's API credit

## 1. Create the key in the Console

1. Sign in to the Claude Console with the organization that holds your plan's API credit.
2. Create a **workspace** just for experiments (for example `arena`) and set a **monthly spend limit** on it. This is your backstop if a script runs away.
3. Create an API key **in that workspace**. Copy it once; the Console won't show it again.

Every key you add (a second $200 key, say) should get its own workspace and spend limit.

## 2. Store it in a file, not your shell

macOS / Linux:

```bash
mkdir -p ~/.config/arena
install -m 600 /dev/null ~/.config/arena/key   # owner-only from the first byte
${EDITOR:-nano} ~/.config/arena/key            # paste, save
```

Windows (PowerShell):

```powershell
$f = "$HOME\.config\arena\key"
New-Item -ItemType File -Force $f | Out-Null
icacls $f /inheritance:r /grant:r "${env:USERNAME}:(F)" | Out-Null   # owner-only before pasting
notepad $f                                                          # paste, save
```

A second key goes in its own file next to the first, for example `~/.config/arena/key2`.

Don't `export` it in `~/.zshrc`, set it in your PowerShell `$PROFILE` or Windows environment settings, and don't paste it into an AI chat or a terminal command; all of those end up in history, transcripts or every process you start.

## 3. Pass it only to the program that needs it

macOS / Linux:

```bash
ARENA_ANTHROPIC_API_KEY="$(cat ~/.config/arena/key)" go run ./your/program
```

Windows (PowerShell):

```powershell
try {
  $env:ARENA_ANTHROPIC_API_KEY = (Get-Content -Raw "$HOME\.config\arena\key").Trim()
  go run ./your/program
} finally { Remove-Item Env:ARENA_ANTHROPIC_API_KEY -ErrorAction SilentlyContinue }
```

Use a **dedicated variable name** (`ARENA_ANTHROPIC_API_KEY`), not `ANTHROPIC_API_KEY`. A dedicated name means:

- other tools on your machine that read `ANTHROPIC_API_KEY` never pick up the experiments key by accident;
- the experiments program can't silently fall back to some other credential.

Full walkthrough, checks and troubleshooting for each OS: [`local-setup.md`](local-setup.md).

## 4. Make the SDK use only that key

The official SDKs look for credentials in several places by default. In the Go SDK, `anthropic.NewClient()` reads, in order: `ANTHROPIC_API_KEY`, `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_PROFILE`, workload-identity variables, and a default profile on disk. It also reads `ANTHROPIC_BASE_URL` and `ANTHROPIC_CUSTOM_HEADERS`.

To use only your explicit key:

```go
client := anthropic.NewClient(
    option.WithoutEnvironmentDefaults(), // ignore every env var and profile
    option.WithHTTPClient(&http.Client{}),
    option.WithAPIKey(key),              // from os.Getenv("ARENA_ANTHROPIC_API_KEY")
    option.WithMaxRetries(0),            // optional: if your code owns retries
)
```

`examples/hello-credits` does exactly this, and its test proves the environment is ignored.

## 5. Never log it

- Don't put the key in structs you log or serialize.
- Clear the local variable after building the client.
- Before committing, search your output directories for `sk-ant` (`grep -r sk-ant arena-data/`, or `Select-String -Path arena-data\* -Pattern sk-ant` on Windows).
