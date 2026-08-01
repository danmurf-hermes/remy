# Telegram Setup Guide

Remy can be used via Telegram, allowing you to chat from your phone while sharing the same memory and context as the desktop GUI.

---

## Prerequisites

- Remy installed and configured
- A Telegram account
- Ollama running locally

---

## Step 1: Create a Telegram Bot

1. Open Telegram and search for **@BotFather** (the official bot for creating bots)
2. Start a chat and send: `/newbot`
3. Follow the prompts:
   - **Name**: Choose a display name (e.g., `My Remy Bot`)
   - **Username**: Choose a unique username ending in `bot` (e.g., `my_remy_bot`)
4. BotFather will give you an **API token** — save it securely. It looks like:
   ```
   1234567890:ABCdefGHIjklmNOPqrstUVwxyz-1234567
   ```

> **⚠️ Security**: Never share your bot token. Anyone with the token can control your bot.

---

## Step 2: Configure Remy

Edit your Remy config file at `~/.remy/config.json`:

```json
{
  "interfaces": {
    "telegram": {
      "enabled": true,
      "bot_token": "1234567890:ABCdefGHIjklmNOPqrstUVwxyz-1234567",
      "allowed_users": ["your_telegram_username"]
    }
  }
}
```

### Configuration Options

| Field | Required | Description |
|-------|----------|-------------|
| `enabled` | Yes | Set to `true` to enable the Telegram interface |
| `bot_token` | Yes | The API token from BotFather |
| `allowed_users` | No | List of Telegram usernames allowed to use the bot. If empty, anyone who finds your bot can use it. |

---

## Step 3: Run Remy in Daemon Mode

```bash
remy --daemon
```

This starts Remy in headless mode — no GUI, just the Telegram bot and scheduler running in the background.

You should see output like:
```
Remy dev starting...
Telegram interface started (bot token: 1234...4567)
Running in daemon mode (no GUI)
```

---

## Step 4: Start Chatting

1. Open Telegram and find your bot by its username
2. Send `/start` to begin
3. Start chatting! Remy will respond with full memory context

### Available Commands

| Command | Description |
|---------|-------------|
| `/start` | Start a conversation |
| `/help` | Show available commands |
| `/status` | Check if Remy is running and connected |

---

## How It Works

- Remy uses **long-polling** to check for new messages from Telegram (no webhooks needed)
- Messages from Telegram are stored with `interface: "telegram"` in the database
- All memory (facts, episodes, scratchpad) is **shared across interfaces** — messages from Telegram appear in the desktop GUI history and vice versa
- The bot shows a **typing indicator** while generating a response
- If the bot token is invalid or Telegram is unreachable, Remy logs the error and continues without the Telegram interface

---

## Troubleshooting

### Bot doesn't respond

1. Check that Remy is running (`remy --daemon`)
2. Verify the bot token is correct in `~/.remy/config.json`
3. Check Remy's logs for errors
4. Make sure Ollama is running and accessible

### "Unauthorized" error

If you set `allowed_users`, make sure your Telegram username is in the list. Usernames are case-insensitive.

### Bot stops responding after a while

The long-polling loop automatically reconnects. Check that:
- Ollama is still running
- Your machine hasn't gone to sleep
- The network connection is stable
