package channel

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

func Register(s *server.MCPServer) {
	registerBasic(s)
	registerAdvanced(s)
}

func registerBasic(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("get_channel",
			mcp.WithDescription("Get detailed information about a single channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			c, err := bot.Channel(channelID)
			if err != nil {
				return tools.Error(err.Error())
			}
			overwrites := make([]map[string]any, 0, len(c.PermissionOverwrites))
			for _, o := range c.PermissionOverwrites {
				otype := "role"
				if o.Type == discordgo.PermissionOverwriteTypeMember {
					otype = "member"
				}
				overwrites = append(overwrites, map[string]any{
					"id": o.ID, "type": otype,
					"allow": strconv.FormatInt(o.Allow, 10), "allow_names": tools.DescribePermissions(o.Allow),
					"deny": strconv.FormatInt(o.Deny, 10), "deny_names": tools.DescribePermissions(o.Deny),
				})
			}
			return tools.JSON(map[string]any{
				"id": c.ID, "name": c.Name, "type": int(c.Type),
				"parent_id": c.ParentID, "position": c.Position, "topic": c.Topic,
				"nsfw": c.NSFW, "rate_limit_per_user": c.RateLimitPerUser,
				"bitrate": c.Bitrate, "user_limit": c.UserLimit,
				"permission_overwrites": overwrites,
			})
		},
	)

	s.AddTool(
		mcp.NewTool("list_channels",
			mcp.WithDescription("List all channels in the server with their types, categories, and positions"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channels, err := bot.GuildChannels(guildID)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Type     int    `json:"type"`
				ParentID string `json:"parent_id,omitempty"`
				Position int    `json:"position"`
				Topic    string `json:"topic,omitempty"`
			}
			result := make([]entry, 0, len(channels))
			for _, c := range channels {
				result = append(result, entry{
					ID: c.ID, Name: c.Name, Type: int(c.Type),
					ParentID: c.ParentID, Position: c.Position, Topic: c.Topic,
				})
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("create_text_channel",
			mcp.WithDescription("Create a new text channel"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name (lowercase, no spaces, use hyphens)")),
			mcp.WithString("category_id", mcp.Description("Parent category ID")),
			mcp.WithString("topic", mcp.Description("Channel topic")),
			mcp.WithBoolean("nsfw", mcp.Description("Whether the channel is NSFW")),
			mcp.WithNumber("rate_limit", mcp.Description("Slowmode in seconds (0-21600)")),
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
			data := discordgo.GuildChannelCreateData{
				Name: name, Type: discordgo.ChannelTypeGuildText,
				ParentID: req.GetString("category_id", ""),
				Topic:    req.GetString("topic", ""),
				NSFW:     req.GetBool("nsfw", false),
			}
			if rl := req.GetFloat("rate_limit", 0); rl > 0 {
				data.RateLimitPerUser = int(rl)
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("create_voice_channel",
			mcp.WithDescription("Create a new voice channel"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("category_id", mcp.Description("Parent category ID")),
			mcp.WithNumber("bitrate", mcp.Description("Bitrate in bits (8000-384000)")),
			mcp.WithNumber("user_limit", mcp.Description("Max users (0 = unlimited, max 99)")),
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
			data := discordgo.GuildChannelCreateData{
				Name: name, Type: discordgo.ChannelTypeGuildVoice,
				ParentID: req.GetString("category_id", ""),
				Bitrate:  int(req.GetFloat("bitrate", 64000)),
			}
			if ul := req.GetFloat("user_limit", 0); ul > 0 {
				data.UserLimit = int(ul)
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("create_category",
			mcp.WithDescription("Create a new channel category"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Category name")),
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
			ch, err := bot.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
				Name: name, Type: discordgo.ChannelTypeGuildCategory,
			})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("edit_channel",
			mcp.WithDescription("Edit an existing channel's properties"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID to edit")),
			mcp.WithString("name", mcp.Description("New channel name")),
			mcp.WithString("topic", mcp.Description("New channel topic")),
			mcp.WithBoolean("nsfw", mcp.Description("Whether NSFW")),
			mcp.WithNumber("rate_limit", mcp.Description("Slowmode in seconds")),
			mcp.WithString("category_id", mcp.Description("Move to category (empty string to remove)")),
			mcp.WithNumber("position", mcp.Description("Channel position")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			edit := &discordgo.ChannelEdit{}
			if v := req.GetString("name", ""); v != "" {
				edit.Name = v
			}
			if v := req.GetString("topic", ""); v != "" {
				edit.Topic = v
			}
			args := req.GetArguments()
			if _, ok := args["nsfw"]; ok {
				b := req.GetBool("nsfw", false)
				edit.NSFW = &b
			}
			if _, ok := args["rate_limit"]; ok {
				rl := int(req.GetFloat("rate_limit", 0))
				edit.RateLimitPerUser = &rl
			}
			if v, ok := args["category_id"]; ok {
				if s, ok := v.(string); ok {
					edit.ParentID = s
				}
			}
			if _, ok := args["position"]; ok {
				pos := int(req.GetFloat("position", 0))
				edit.Position = &pos
			}
			ch, err := bot.ChannelEditComplex(channelID, edit)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("delete_channel",
			mcp.WithDescription("Delete a channel (irreversible)"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID to delete")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if _, err = bot.ChannelDelete(channelID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("channel deleted"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("set_channel_permissions",
			mcp.WithDescription("Set permission overrides for a role or user on a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("target_id", mcp.Required(), mcp.Description("Role ID or User ID")),
			mcp.WithString("target_type", mcp.Required(), mcp.Description("'role' or 'member'")),
			mcp.WithString("allow", mcp.Description("Allowed permissions as bitfield string")),
			mcp.WithString("deny", mcp.Description("Denied permissions as bitfield string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			targetID, err := req.RequireString("target_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			targetType, err := req.RequireString("target_type")
			if err != nil {
				return tools.Error(err.Error())
			}
			dType := discordgo.PermissionOverwriteTypeMember
			if targetType == "role" {
				dType = discordgo.PermissionOverwriteTypeRole
			}
			allow, _ := strconv.ParseInt(req.GetString("allow", "0"), 10, 64)
			deny, _ := strconv.ParseInt(req.GetString("deny", "0"), 10, 64)

			if err = bot.ChannelPermissionSet(channelID, targetID, dType, allow, deny); err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]interface{}{
				"channel_id": channelID, "target_id": targetID, "target_type": targetType,
				"allow": allow, "allow_names": tools.DescribePermissions(allow),
				"deny": deny, "deny_names": tools.DescribePermissions(deny),
			})
		},
	)

	s.AddTool(
		mcp.NewTool("sync_category_permissions",
			mcp.WithDescription("Set permission overrides on a category AND sync to all child channels"),
			mcp.WithString("category_id", mcp.Required(), mcp.Description("Category channel ID")),
			mcp.WithString("target_id", mcp.Required(), mcp.Description("Role ID or User ID")),
			mcp.WithString("target_type", mcp.Required(), mcp.Description("'role' or 'member'")),
			mcp.WithString("allow", mcp.Description("Allowed permissions as bitfield string")),
			mcp.WithString("deny", mcp.Description("Denied permissions as bitfield string")),
			mcp.WithBoolean("force", mcp.Description("Overwrite existing custom overrides on child channels")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			categoryID, err := req.RequireString("category_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			targetID, err := req.RequireString("target_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			targetType, err := req.RequireString("target_type")
			if err != nil {
				return tools.Error(err.Error())
			}
			cat, err := bot.Channel(categoryID)
			if err != nil {
				return tools.Errorf("category not found: %v", err)
			}
			if cat.Type != discordgo.ChannelTypeGuildCategory {
				return tools.Error("channel is not a category")
			}
			dType := discordgo.PermissionOverwriteTypeMember
			if targetType == "role" {
				dType = discordgo.PermissionOverwriteTypeRole
			}
			allow, _ := strconv.ParseInt(req.GetString("allow", "0"), 10, 64)
			deny, _ := strconv.ParseInt(req.GetString("deny", "0"), 10, 64)
			force := req.GetBool("force", false)

			if err = bot.ChannelPermissionSet(categoryID, targetID, dType, allow, deny); err != nil {
				return tools.Errorf("failed to set category permissions: %v", err)
			}
			channels, err := bot.GuildChannels(guildID)
			if err != nil {
				return tools.Errorf("failed to list channels: %v", err)
			}
			var synced, skipped, failed []string
			for _, ch := range channels {
				if ch.ParentID != categoryID {
					continue
				}
				if !force {
					hasOverride := false
					for _, perm := range ch.PermissionOverwrites {
						if perm.ID == targetID {
							hasOverride = true
							break
						}
					}
					if hasOverride {
						skipped = append(skipped, ch.Name)
						continue
					}
				}
				if err := bot.ChannelPermissionSet(ch.ID, targetID, dType, allow, deny); err != nil {
					failed = append(failed, fmt.Sprintf("%s: %s", ch.Name, err.Error()))
				} else {
					synced = append(synced, ch.Name)
				}
			}
			return tools.JSON(map[string]interface{}{
				"category_name": cat.Name, "target_type": targetType,
				"allow_names": tools.DescribePermissions(allow),
				"deny_names":  tools.DescribePermissions(deny),
				"synced":      synced, "skipped": skipped, "failed": failed,
			})
		},
	)
}
