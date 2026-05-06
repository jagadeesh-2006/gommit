# gommit ⚡

> AI-powered git commit message generator — single binary, no Node required

![demo](assets/demo.gif)

---

## Why gommit?

Every AI commit tool is built in Node.js — which means installing npm, a runtime, and dozens of dependencies just to get a CLI utility.

**gommit is different:**
- Single binary — no runtime, no npm, no dependencies
- Bring your own API key — your code never leaves your machine
- Supports multiple AI providers — Groq (free), Anthropic, OpenAI
- Works on Windows, Mac, and Linux

---

## Install

### Option 1 — go install *(recommended)*
Requires [Go](https://golang.org/dl/) installed.

```bash
go install github.com/jagadeesh-2006/gommit@latest
```

### Option 2 — Build from Source

```bash
# 1. clone the repo
git clone https://github.com/jagadeesh-2006/gommit.git
cd gommit

# 2. install dependencies
go mod tidy

# 3. build binary
go build -o gommit .

# 4. move to PATH (mac/linux)
sudo mv gommit /usr/local/bin/

# windows — move to go bin
copy gommit.exe %USERPROFILE%\go\bin\
```

---

## Quick Start

### Step 1 — Run setup wizard

```bash
gommit init
```

```
Enter your AI provider (anthropic, groq, openai): groq
Enter your API key: gsk_xxxxxxxxxxxx
Fetching models...
Available models:
1. llama-3.3-70b-versatile
2. llama-3.1-8b-instant
3. mixtral-8x7b-32768
Select a model number: 1
Commit style (conventional, simple, emoji): conventional
Custom prompt (optional, press Enter to skip):
✅ Config saved. Run `gommit run` in any repo.
```

### Step 2 — Stage your changes

```bash
git add .
```

### Step 3 — Generate and commit

```bash
gommit run
```

```
Suggested commit: feat(auth): add JWT token refresh logic

[y] commit  [e] edit  [r] regenerate  [n] cancel
```

- `y` — accept and commit
- `e` — edit the message then commit
- `r` — regenerate a new message
- `n` — cancel without committing

---

## Commands

| Command | Alias | Description |
|---|---|---|
| `gommit init` | `i` | First time setup wizard |
| `gommit run` | `r` | Generate AI commit message and commit |
| `gommit config` | `cfg` | View current configuration |
| `gommit update` | `u` | Update any configuration value |
| `gommit prompt` | `p` | Set custom prompt instructions |
| `gommit undo` | `un` | Undo last commit, keep changes staged |
| `gommit uninstall` | `remove` | Remove gommit configuration |

---

### `gommit init`
First time setup. Prompts for provider, API key, model, commit style, and optional custom prompt.

```bash
gommit init
```

---

### `gommit run`
Reads your staged diff, sends to AI, suggests a commit message, and commits on approval.

```bash
git add .
gommit run
```

---

### `gommit config`
View your current saved configuration.

```bash
gommit config

# output:
# Current Configuration:
#   Provider:      groq
#   Model:         llama-3.3-70b-versatile
#   API Key:       gsk_xxxx****
#   Commit Style:  conventional
#   Custom Prompt: none
```

---

### `gommit update`
Update any config value without re-running init. Auto-fetches models when switching provider.

```bash
gommit update

# Select the setting you want to update:
# 1. Provider + model + API key
# 2. Model
# 3. API Key
# 4. Commit Style
# 5. Custom Prompt
```

---

### `gommit prompt`
Set custom instructions for how you want your commit messages generated.

```bash
gommit prompt

# Current prompt: none
# Enter your prompt instructions: keep messages under 50 chars, focus on why not what
# ✅ Prompt saved.
```

---

### `gommit undo`
Undo your last commit and keep changes staged — ready to recommit.

```bash
gommit undo
# runs: git reset --soft HEAD~1
# your staged changes come back, nothing is lost
```

---

### `gommit uninstall`
Remove gommit configuration from your machine.

```bash
gommit uninstall

# Are you sure you want to remove gommit config? (y/n): y
# ✅ gommit config removed successfully.
# Note: Binary is still installed. To fully remove:
#   Windows:   del %USERPROFILE%\go\bin\gommit.exe
#   Mac/Linux: rm /usr/local/bin/gommit
```

---

## Supported Providers

| Provider | Free Tier | Models |
|---|---|---|
| [Groq](https://console.groq.com) | ✅ Free | Llama 3.3, Mixtral, Gemma |
| [Anthropic](https://console.anthropic.com) | ❌ Paid | Claude 3.5, Claude 3 |
| [OpenAI](https://platform.openai.com/api-keys) | ❌ Paid | GPT-4o, GPT-4 Turbo |

> **New to this?** Start with Groq — completely free, no credit card required.
> Get your key at [console.groq.com](https://console.groq.com)

---

## Commit Styles

| Style | Example Output |
|---|---|
| `conventional` | `feat(auth): add JWT token refresh logic` |
| `simple` | `add JWT token refresh logic` |
| `emoji` | `✨ add JWT token refresh logic` |
| `any` | define your own style via custom prompt |

---

## Configuration

Stored at `~/.gommit/config.json` — never shared, never transmitted beyond your chosen AI provider.

```json
{
  "version": "1",
  "provider": "groq",
  "model": "llama-3.3-70b-versatile",
  "api_key": "gsk_xxxxxxxxxxxx",
  "commit_style": "conventional",
  "custom_prompt": ""
}
```

---

## Troubleshooting

**"run gommit init first"**
→ No config found. Run `gommit init` to set up.

**"no staged changes found"**
→ Stage your files first with `git add .`

**"invalid provider"**
→ Supported providers: `groq`, `anthropic`, `openai`

**"error fetching models"**
→ API key may be invalid. Run `gommit update` → option 3 to fix it.

**"rate limit exceeded"**
→ Wait a moment or switch to a different provider with `gommit update`.

**"model returned empty response"**
→ Selected model may not support chat. Run `gommit update` → option 2 and pick a different model .

**"could not reach provider"**
→ Check your internet connection or try again later.

---

## Security

- API keys stored at `~/.gommit/config.json` with `0600` permissions — owner read only
- Git diffs are sent **only** to your chosen AI provider — never to any gommit server
- gommit has no backend — everything runs entirely on your machine
- Sensitive data in diffs (passwords, tokens, keys) triggers a warning before sending
- Prompts are open source — what you see in code is exactly what gets sent to the AI

---

## Uninstall

```bash
# step 1 — remove config
gommit uninstall

# step 2 — remove binary
# windows
del %USERPROFILE%\go\bin\gommit.exe

# mac/linux
rm /usr/local/bin/gommit
```

---

## Contributing

PRs are welcome. Please open an issue first for major changes.

```bash
# 1. fork the repo

# 2. create a feature branch
git checkout -b feature/your-feature

# 3. make your changes and commit
gommit run

# 4. push and open a PR
git push origin feature/your-feature
```

---

## License

MIT — see [LICENSE](LICENSE) for details.

---

**Built with [Cobra](https://github.com/spf13/cobra) · Powered by Groq, Anthropic, OpenAI**