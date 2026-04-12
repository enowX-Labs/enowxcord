package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerServerTools(s *server.MCPServer, d *Discord) {
	bot := d.Session
	guildID := d.GuildID

	s.AddTool(
		mcp.NewTool("get_server_info",
			mcp.WithDescription("Get detailed server information including name, icon, member count, boost level, features"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			guild, err := bot.GuildWithCounts(guildID)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]interface{}{
				"id":                         guild.ID,
				"name":                       guild.Name,
				"description":                guild.Description,
				"member_count":               guild.ApproximateMemberCount,
				"online_count":               guild.ApproximatePresenceCount,
				"premium_tier":               guild.PremiumTier,
				"premium_subscription_count": guild.PremiumSubscriptionCount,
				"features":                   guild.Features,
			})
		},
	)

	s.AddTool(
		mcp.NewTool("edit_server",
			mcp.WithDescription("Edit server settings (name, description)"),
			mcp.WithString("name", mcp.Description("New server name")),
			mcp.WithString("description", mcp.Description("New server description")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			gp := discordgo.GuildParams{}
			if v := req.GetString("name", ""); v != "" {
				gp.Name = v
			}
			if v := req.GetString("description", ""); v != "" {
				gp.Description = v
			}
			guild, err := bot.GuildEdit(guildID, &gp)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"name": guild.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("list_emojis",
			mcp.WithDescription("List all custom emojis in the server"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			emojis, err := bot.GuildEmojis(guildID)
			if err != nil {
				return toolError(err.Error())
			}
			type em struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Animated bool   `json:"animated"`
			}
			result := make([]em, 0, len(emojis))
			for _, emoji := range emojis {
				result = append(result, em{ID: emoji.ID, Name: emoji.Name, Animated: emoji.Animated})
			}
			return resultJSON(result)
		},
	)
}
