package guild

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

func Register(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("get_server_info",
			mcp.WithDescription("Get detailed server information including name, icon, member count, boost level, features"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			g, err := bot.GuildWithCounts(guildID)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]interface{}{
				"id": g.ID, "name": g.Name, "description": g.Description,
				"member_count": g.ApproximateMemberCount, "online_count": g.ApproximatePresenceCount,
				"premium_tier": g.PremiumTier, "premium_subscription_count": g.PremiumSubscriptionCount,
				"features": g.Features, "vanity_url_code": g.VanityURLCode, "icon": g.Icon,
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
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
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
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
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

	s.AddTool(
		mcp.NewTool("create_emoji",
			mcp.WithDescription("Upload a custom emoji to the server"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Emoji name (alphanumeric and underscores)")),
			mcp.WithString("image", mcp.Required(), mcp.Description("Image as a data URI (data:image/png;base64,...) or a raw base64 string. Must be <256KB.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			name, err := req.RequireString("name")
			if err != nil {
				return tools.Error(err.Error())
			}
			image, err := req.RequireString("image")
			if err != nil {
				return tools.Error(err.Error())
			}
			// discordgo expects a full data URI; wrap a bare base64 string as PNG.
			if !strings.HasPrefix(image, "data:") {
				image = "data:image/png;base64," + image
			}
			e, err := bot.GuildEmojiCreate(guildID, &discordgo.EmojiParams{Name: name, Image: image})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": e.ID, "name": e.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("delete_emoji",
			mcp.WithDescription("Delete a custom emoji from the server"),
			mcp.WithString("emoji_id", mcp.Required(), mcp.Description("Emoji ID to delete")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			emojiID, err := req.RequireString("emoji_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.GuildEmojiDelete(guildID, emojiID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("emoji deleted"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("set_server_icon",
			mcp.WithDescription("Set the server icon"),
			mcp.WithString("image", mcp.Required(), mcp.Description("Image as a data URI (data:image/png;base64,...) or raw base64 string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			image, err := req.RequireString("image")
			if err != nil {
				return tools.Error(err.Error())
			}
			if !strings.HasPrefix(image, "data:") {
				image = "data:image/png;base64," + image
			}
			if _, err = bot.GuildEdit(guildID, &discordgo.GuildParams{Icon: image}); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("server icon updated"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_integrations",
			mcp.WithDescription("List server integrations (bots, Twitch/YouTube links, etc.)"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			integrations, err := bot.GuildIntegrations(guildID)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Type    string `json:"type"`
				Enabled bool   `json:"enabled"`
			}
			result := make([]entry, 0, len(integrations))
			for _, i := range integrations {
				result = append(result, entry{ID: i.ID, Name: i.Name, Type: i.Type, Enabled: i.Enabled})
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("get_audit_log",
			mcp.WithDescription("Get recent audit log entries (administrative actions)"),
			mcp.WithNumber("limit", mcp.Description("Max entries (1-100, default 50)")),
			mcp.WithString("user_id", mcp.Description("Filter by the user who performed actions")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			limit := int(req.GetFloat("limit", 50))
			if limit < 1 || limit > 100 {
				limit = 50
			}
			log, err := bot.GuildAuditLog(guildID, req.GetString("user_id", ""), "", 0, limit)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				ID         string `json:"id"`
				UserID     string `json:"user_id"`
				ActionType int    `json:"action_type"`
				TargetID   string `json:"target_id,omitempty"`
				Reason     string `json:"reason,omitempty"`
			}
			result := make([]entry, 0, len(log.AuditLogEntries))
			for _, a := range log.AuditLogEntries {
				e := entry{ID: a.ID, UserID: a.UserID, TargetID: a.TargetID, Reason: a.Reason}
				if a.ActionType != nil {
					e.ActionType = int(*a.ActionType)
				}
				result = append(result, e)
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("prune_members",
			mcp.WithDescription("Remove members who have been inactive and have no roles. Use dry_run to preview the count first."),
			mcp.WithNumber("days", mcp.Required(), mcp.Description("Inactivity threshold in days (1-30)")),
			mcp.WithBoolean("dry_run", mcp.Description("If true, only return the count that would be pruned without removing anyone")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			days := uint32(req.GetFloat("days", 0))
			if days < 1 || days > 30 {
				return tools.Error("days must be between 1 and 30")
			}
			if req.GetBool("dry_run", false) {
				count, err := bot.GuildPruneCount(guildID, days)
				if err != nil {
					return tools.Error(err.Error())
				}
				return tools.JSON(map[string]any{"dry_run": true, "would_prune": count})
			}
			count, err := bot.GuildPrune(guildID, days)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]any{"pruned": count})
		},
	)
}
