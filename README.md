# gommit ⚡

> AI-powered git commit message generator — single binary, no node required


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

# windows — run directly from folder
./gommit.exe init
```

---

## Quick Start

### Step 1 — Run setup wizard

```bash
gommit init
```

You will be prompted to:

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

### Step 3 — Generate commit message

```bash
gommit run
```

```
Suggested commit: feat(auth): add JWT token refresh logic

[y] commit  [e] edit  [r] regenerate  [n] cancel
```

Choose your action:
- `y` — accept and commit
- `e` — edit the message then commit
- `r` — regenerate a new message
- `n` — cancel

---

## Commands

| Command | Shortcut | Description |
|---|---|---|
| `gommit init` | `gommit i` | First time setup wizard |
| `gommit run` | `gommit r` | Generate and commit |
| `gommit config` | `gommit cfg` | View current configuration |
| `gommit update` | `gommit u` | Update configuration values |
| `gommit prompt` | `gommit p` | Set custom prompt instructions |

---

## Supported Providers

| Provider | Free Tier | Models |
|---|---|---|
| [Groq](https://console.groq.com) | ✅ Free | Llama 3.3, Mixtral |
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
| `any` | your custom style via prompt |

---

## Custom Prompt

Give gommit specific instructions for how you want your commit messages:

```bash
gommit prompt
```

```
Current prompt: none
Enter your prompt instructions: keep messages under 50 chars, focus on why not what
✅ Prompt saved.
```

This gets appended to every AI request as additional instructions.

---

## Update Configuration

Change any setting without re-running init:

```bash
gommit update
```

```
Select the setting you want to update:
1. Provider + model + API key
2. Model
3. API Key
4. Commit Style
5. Custom Prompt
```

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

---

## Security

- API keys stored at `~/.gommit/config.json` with `0600` permissions (owner read only)
- Git diffs are only sent to your chosen AI provider
- No backend — everything runs on your machine

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