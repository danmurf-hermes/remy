// Package telegram implements a Telegram bot interface for Remy using
// long-polling via the go-telegram-bot-api library. It receives messages
// from Telegram, routes them through the agent, and sends responses back.
package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/memory"
)

// AgentService defines the subset of agent.Agent methods the Telegram
// interface needs, making it easy to mock in tests.
type AgentService interface {
	HandleMessage(ctx context.Context, userMsg string) (*memory.Message, error)
}

// Store defines the subset of memory.Store methods the Telegram interface
// needs, making it easy to mock in tests.
type Store interface {
	SaveMessage(ctx context.Context, msg *memory.Message) error
}

// Interface is the Telegram bot interface. It connects to Telegram via
// long-polling, receives messages, sends them to the agent, and returns
// responses to the chat.
type Interface struct {
	bot     *tgbotapi.BotAPI
	agent   AgentService
	store   Store
	cfg     *config.TelegramConfig
	stopCh  chan struct{}
	wg      sync.WaitGroup
	started bool
	mu      sync.Mutex
}

// New creates a new Telegram interface with the given dependencies.
func New(agent AgentService, store Store, cfg *config.TelegramConfig) *Interface {
	return &Interface{
		agent:  agent,
		store:  store,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

// Start connects to Telegram and begins polling for updates. It blocks
// until the bot is connected, then returns. Call Stop to shut down.
func (t *Interface) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return fmt.Errorf("telegram interface already started")
	}
	t.mu.Unlock()

	if t.cfg.BotToken == "" {
		return fmt.Errorf("telegram bot token is empty")
	}

	bot, err := tgbotapi.NewBotAPI(t.cfg.BotToken)
	if err != nil {
		return fmt.Errorf("creating telegram bot: %w", err)
	}

	t.bot = bot
	bot.Debug = false

	user, err := bot.GetMe()
	if err != nil {
		return fmt.Errorf("getting bot info: %w", err)
	}
	log.Printf("Telegram bot connected: @%s (ID: %d)", user.UserName, user.ID)

	t.mu.Lock()
	t.started = true
	t.mu.Unlock()

	t.wg.Add(1)
	go t.pollLoop(ctx)

	return nil
}

// Stop gracefully shuts down the Telegram interface, stopping the poll
// loop and waiting for in-flight message processing to complete.
func (t *Interface) Stop() {
	t.mu.Lock()
	if !t.started {
		t.mu.Unlock()
		return
	}
	t.started = false
	t.mu.Unlock()

	close(t.stopCh)
	t.wg.Wait()

	if t.bot != nil {
		t.bot.StopReceivingUpdates()
	}
	log.Println("Telegram interface stopped")
}

// Send sends a text message to the specified chat ID.
func (t *Interface) Send(chatID int64, text string) error {
	if t.bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := t.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sending telegram message: %w", err)
	}
	return nil
}

// pollLoop is the main polling loop that receives updates from Telegram
// and processes incoming messages.
func (t *Interface) pollLoop(ctx context.Context) {
	defer t.wg.Done()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for {
		select {
		case <-t.stopCh:
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Message == nil {
				continue
			}
			t.handleUpdate(ctx, update)
		}
	}
}

// handleUpdate processes a single Telegram update containing a message.
func (t *Interface) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	msg := update.Message
	if msg == nil {
		return
	}

	chatID := msg.Chat.ID
	userID := msg.From.ID
	text := strings.TrimSpace(msg.Text)

	// Ignore empty messages and non-text messages
	if text == "" {
		return
	}

	// Check if user is allowed
	if !t.isUserAllowed(userID) {
		log.Printf("Telegram: blocked message from unauthorized user %d", userID)
		t.sendWithRetry(chatID, "Sorry, you are not authorized to use this bot.")
		return
	}

	// Handle commands
	if msg.IsCommand() {
		t.handleCommand(ctx, chatID, userID, msg.Command(), msg.CommandArguments())
		return
	}

	log.Printf("Telegram: message from user %d in chat %d: %.50s", userID, chatID, text)

	// Send typing indicator
	t.sendChatAction(chatID)

	// Process through agent
	agentCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	response, err := t.agent.HandleMessage(agentCtx, text)
	if err != nil {
		log.Printf("Telegram: agent error for user %d: %v", userID, err)
		t.sendWithRetry(chatID, "Sorry, I encountered an error processing your message.")
		return
	}

	if response == nil || response.Content == "" {
		return
	}

	t.sendWithRetry(chatID, response.Content)
}

// handleCommand processes Telegram bot commands.
func (t *Interface) handleCommand(ctx context.Context, chatID int64, userID int64, command, args string) {
	switch command {
	case "start":
		t.sendWithRetry(chatID, "Hello! I'm Remy, your personal AI assistant. Send me a message and I'll respond.")
	case "help":
		t.sendWithRetry(chatID, "I'm Remy, your personal AI assistant. Just send me a message and I'll help you out.\n\nCommands:\n/start - Start the bot\n/help - Show this help message\n/status - Check bot status")
	case "status":
		t.sendWithRetry(chatID, "Remy is online and ready to help!")
	default:
		t.sendWithRetry(chatID, fmt.Sprintf("Unknown command: /%s. Use /help to see available commands.", command))
	}
}

// isUserAllowed checks if the given Telegram user ID is in the allowed users list.
// If the allowed users list is empty, all users are allowed.
func (t *Interface) isUserAllowed(userID int64) bool {
	if len(t.cfg.AllowedUsers) == 0 {
		return true
	}
	uidStr := strconv.FormatInt(userID, 10)
	for _, allowed := range t.cfg.AllowedUsers {
		if allowed == uidStr {
			return true
		}
	}
	return false
}

// sendChatAction sends a typing indicator to the chat.
func (t *Interface) sendChatAction(chatID int64) {
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	_, _ = t.bot.Request(action)
}

// sendWithRetry sends a message with a single retry on failure.
func (t *Interface) sendWithRetry(chatID int64, text string) {
	err := t.Send(chatID, text)
	if err != nil {
		log.Printf("Telegram: send error (retrying): %v", err)
		time.Sleep(500 * time.Millisecond)
		_ = t.Send(chatID, text)
	}
}
