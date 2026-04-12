package discord

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Client struct {
	Session *discordgo.Session
	GuildID string
}

func New() (*Client, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}

	guildID := os.Getenv("GUILD_ID")
	if guildID == "" {
		return nil, fmt.Errorf("GUILD_ID environment variable is required")
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
