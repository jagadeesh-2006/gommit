# gommit ⚡

> AI-powered git commit message generator — single binary, no Node required


---

## Why gommit?

Every AI commit tool is built in Node.js — which means installing npm, a runtime, and 50 dependencies just to get a CLI util.

gommit is a single binary. Install it once, works everywhere.

---

## Install

### Option 1 — go install
```bash
go install github.com/jagadeesh-2006/gommit@latest
```

### Option 2 — From Source
```bash
git clone https://github.com/jagadeesh-2006/gommit.git
cd gommit
go build -o gommit .

# mac/linux
sudo mv gommit /usr/local/bin/

# windows
./gommit.exe init
```

---

## Setup

Run once to configure your provider and API key:

```bash
gommit init
```

```
Which provider? (anthropic, groq, openai): groq
Enter API key: gsk_xxxx
Fetching models...
1. llama-3.3-70b-versatile
2. llama-3.1-8b-instant
3. mixtral-8x7b-32768
Select model: 1
Commit style (conventional, simple, emoji): conventional
✅ Config saved. Run `gommit run` in any repo.
```

---

## Usage

```bash
git add .
gommit run
```

```
Suggested commit: feat(auth): add JWT token refresh logic

[y] commit  [e] edit  [r] regenerate  [n] cancel
```

---

## Supported Providers

| Provider | Free Tier | Models |
|---|---|---|
| Groq | ✅ Yes | Llama 3, Mixtral |
| Anthropic | ❌ Paid | Claude 3.5, Claude 3 |
| OpenAI | ❌ Paid | GPT-4o, GPT-4 turbo |

> New to this? Start with Groq — completely free. Get your key at console.groq.com

---

## Commit Styles

| Style | Example |
|---|---|
| conventional | `feat(auth): add login page` |
| simple | `add login page` |
| emoji | `✨ add login page` |

---

## Config

Stored at `~/.gommit/config.json` — never shared, never leaves your machine.

```json
{
  "version": "1",
  "provider": "groq",
  "model": "llama-3.3-70b-versatile",
  "api_key": "gsk_xxxx",
  "commit_style": "conventional"
}
```

---

## Troubleshooting

**"run gommit init first"**
→ Config not found. Run `gommit init`

**"no staged changes found"**
→ Stage your files first with `git add .`

**"invalid provider"**
→ Only `groq`, `anthropic`, `openai` supported right now

---

## Contributing

PRs welcome. Open an issue first for big changes.

1. Fork the repo
2. Create a branch `git checkout -b feature/your-feature`
3. Commit your changes
4. Open a PR

---

## License

MIT