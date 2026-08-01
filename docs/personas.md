# Personas Guide

Personas let you customize Remy's personality, behavior, and even the underlying LLM model. Each persona is a Markdown file with YAML frontmatter stored in `~/.remy/personas/`.

---

## Creating a Persona

Personas are stored as `.md` files in `~/.remy/personas/`. The filename (without `.md`) becomes the persona name.

### Minimal Persona

```markdown
---
name: helper
---

You are a helpful assistant who answers questions clearly and concisely.
```

### Full Persona with Model Override

```markdown
---
name: code-wizard
provider: ollama
model: codellama:13b
temperature: 0.2
max_tokens: 4096
---

You are an expert programming assistant. You write clean, idiomatic code and explain your reasoning. You prefer simple solutions over complex ones. When suggesting code, always include the full file path as a comment.
```

---

## Persona Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `name` | Yes | — | Display name for the persona |
| `provider` | No | Config default | LLM provider to use (e.g., `ollama`) |
| `model` | No | Config default | Model name (e.g., `llama3.1:8b`, `codellama:13b`) |
| `temperature` | No | Config default | Response creativity (0.0–2.0). Lower = more deterministic |
| `max_tokens` | No | Config default | Maximum response length in tokens |

### Field Details

**`name`**: Used for display in the GUI and for switching personas via conversation ("switch to code-wizard").

**`provider`**: Overrides the default provider from `~/.remy/config.json`. The provider must be configured in `config.json` under `providers`.

**`model`**: Overrides the default model for the selected provider. Must be a model available on that provider.

**`temperature`**: Controls randomness in responses:
- **0.0–0.3**: Focused, deterministic responses (good for coding, analysis)
- **0.4–0.7**: Balanced, creative responses (good for conversation)
- **0.8–2.0**: Highly creative, sometimes unpredictable (good for brainstorming)

**`max_tokens`**: Maximum number of tokens in the response. Higher values allow longer responses but use more context window.

---

## Persona Body (System Prompt)

The Markdown body after the YAML frontmatter becomes the **system prompt** for that persona. This is the core instruction that defines how Remy behaves.

### Tips for Writing Good System Prompts

- **Be specific**: "You are a friendly assistant" is less effective than "You are a warm, empathetic listener who asks follow-up questions"
- **Set boundaries**: "You never give medical advice. Instead, suggest consulting a doctor."
- **Define tone**: "Use casual, conversational language with occasional emojis"
- **Include examples**: Show the kind of response you want

### Example: Friendly Companion

```markdown
---
name: friend
temperature: 0.8
---

You are a warm, supportive friend who remembers everything about the user's life. You ask about their day, their projects, and their interests. You're genuinely curious and always follow up on things they've mentioned before. Use casual language and occasional emojis. Never be formal or robotic.
```

### Example: Professional Coach

```markdown
---
name: coach
temperature: 0.4
---

You are a professional productivity coach. You help the user stay focused on their goals, break down large tasks, and maintain momentum. You're direct but encouraging. You reference past conversations to track progress. You never make excuses — you help the user find solutions.
```

---

## Switching Personas

### Via Conversation

Just say "switch to [persona name]" in any conversation:

```
User: switch to code-wizard
Remy: Switched to code-wizard persona. I'm now in expert programming mode.
```

### Via GUI

Open the **Personas** tab in the desktop app. Click on any persona to activate it.

### Via Config

Edit `~/.remy/config.json` and set:

```json
{
  "persona": {
    "active": "code-wizard"
  }
}
```

---

## Default Persona

When you run `remy init`, a default persona is created at `~/.remy/personas/default.md`:

```markdown
---
name: default
---

You are Remy, a personal AI assistant. You are helpful, warm, and conversational. You remember past conversations and use that context to provide better responses. You can help with questions, tasks, reminders, and general conversation.
```

You can edit or replace this file at any time.

---

## Persona Studio (GUI)

The desktop app includes a **Persona Studio** tab where you can:

- **Browse** all available personas
- **Edit** persona system prompts with a live preview toggle
- **Configure** model parameters (provider, model, temperature, max_tokens)
- **Create** new personas
- **Activate** any persona with one click

---

## Troubleshooting

### Persona not showing up

- Make sure the file is in `~/.remy/personas/` with a `.md` extension
- Check that the YAML frontmatter is valid (enclosed in `---` delimiters)
- Verify the `name` field is present in the frontmatter

### Model override not working

- Ensure the provider is configured in `~/.remy/config.json`
- Verify the model name is correct and available on the provider
- Check that Ollama (or your provider) is running

### Persona switch not recognized

- The agent looks for patterns like "switch to X", "change to X", "use X", "activate X"
- The persona name must match the filename (without `.md`)
- Names are case-insensitive
