package discord

import (
	"context"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type contextKey string

const (
	sessionKey contextKey = "discord_session"
	guildKey   contextKey = "discord_guild"
)

type Client struct {
	Session *discordgo.Session
	GuildID string
}

func NewFromCredentials(token, guildID string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("discord token is required")
	}
	if guildID == "" {
		return nil, fmt.Errorf("guild ID is required")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildWebhooks |
		discordgo.IntentsGuildInvites |
		discordgo.IntentsGuildEmojis

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("failed to open discord session: %w", err)
	}

	return &Client{Session: session, GuildID: guildID}, nil
}

func (c *Client) Close() {
	if c.Session != nil {
		c.Session.Close()
	}
}

func WithClient(ctx context.Context, client *Client) context.Context {
	ctx = context.WithValue(ctx, sessionKey, client.Session)
	ctx = context.WithValue(ctx, guildKey, client.GuildID)
	return ctx
}

func SessionFromContext(ctx context.Context) *discordgo.Session {
	if v, ok := ctx.Value(sessionKey).(*discordgo.Session); ok {
		return v
	}
	return nil
}

func GuildIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(guildKey).(string); ok {
		return v
	}
	return ""
}

type SessionPool struct {
	mu       sync.RWMutex
	sessions map[string]*Client
}

func NewSessionPool() *SessionPool {
	return &SessionPool{sessions: make(map[string]*Client)}
}

func (p *SessionPool) Get(token, guildID string) (*Client, error) {
	key := token + ":" + guildID

	p.mu.RLock()
	if c, ok := p.sessions[key]; ok {
		p.mu.RUnlock()
		return c, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	if c, ok := p.sessions[key]; ok {
		return c, nil
	}

	c, err := NewFromCredentials(token, guildID)
	if err != nil {
		return nil, err
	}
	p.sessions[key] = c
	return c, nil
}

func (p *SessionPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.sessions {
		c.Close()
	}
	p.sessions = make(map[string]*Client)
}
