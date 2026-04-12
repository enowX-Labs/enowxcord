package guild

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

func Register(s *server.MCPServer, bot *discordgo.Session, guildID string) {
	s.AddTool(
		mcp.NewTool("get_server_info",
			mcp.WithDescription("Get detailed server information including name, icon, member count, boost level, features"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			g, err := bot.GuildWithCounts(guildID)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]interface{}{
				"id": g.ID, "name": g.Name, "description": g.Description,
				"member_count": g.ApproximateMemberCount, "online_count": g.ApproximatePresenceCount,
				"premium_tier": g.PremiumTier, "premium_subscription_count": g.PremiumSubscriptionCount,
				"features": g.Features,
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
			g, err := bot.GuildEdit(guildID, &gp)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"name": g.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("list_emojis",
			mcp.WithDescription("List all custom emojis in the server"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			emojis, err := bot.GuildEmojis(guildID)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Animated bool   `json:"animated"`
			}
			result := make([]entry, 0, len(emojis))
			for _, e := range emojis {
				result = append(result, entry{ID: e.ID, Name: e.Name, Animated: e.Animated})
			}
			return tools.JSON(result)
		},
	)
}
