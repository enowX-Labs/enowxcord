package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Session *discordgo.Session
	GuildID string
}

func NewDiscord() (*Discord, error) {
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

	return &Discord{Session: session, GuildID: guildID}, nil
}

func (d *Discord) Close() {
	if d.Session != nil {
		d.Session.Close()
	}
}
