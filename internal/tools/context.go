package tools

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/enowx/enowxcord/internal/discord"
)

func FromContext(ctx context.Context) (*discordgo.Session, string, *mcp.CallToolResult) {
	bot := discord.SessionFromContext(ctx)
	guildID := discord.GuildIDFromContext(ctx)
	if bot == nil || guildID == "" {
		result := mcp.NewToolResultError("missing X-Discord-Token or X-Guild-ID headers")
		return nil, "", result
	}
	return bot, guildID, nil
}

func BotFromContext(ctx context.Context) (*discordgo.Session, *mcp.CallToolResult) {
	bot := discord.SessionFromContext(ctx)
	if bot == nil {
		result := mcp.NewToolResultError("missing X-Discord-Token header")
		return nil, result
	}
	return bot, nil
}
