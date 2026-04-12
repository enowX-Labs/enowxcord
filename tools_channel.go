package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerChannelTools(s *server.MCPServer, d *Discord) {
	bot := d.Session
	guildID := d.GuildID

	// list_channels
	s.AddTool(
		mcp.NewTool("list_channels",
			mcp.WithDescription("List all channels in the server with their types, categories, and positions"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channels, err := bot.GuildChannels(guildID)
			if err != nil {
				return toolError(err.Error())
			}
			type ch struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Type     int    `json:"type"`
				ParentID string `json:"parent_id,omitempty"`
				Position int    `json:"position"`
				Topic    string `json:"topic,omitempty"`
			}
			result := make([]ch, 0, len(channels))
			for _, c := range channels {
				result = append(result, ch{
					ID:       c.ID,
					Name:     c.Name,
					Type:     int(c.Type),
					ParentID: c.ParentID,
					Position: c.Position,
					Topic:    c.Topic,
				})
			}
			return resultJSON(result)
		},
	)

	// create_text_channel
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
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error())
			}
			data := discordgo.GuildChannelCreateData{
				Name:     name,
				Type:     discordgo.ChannelTypeGuildText,
				ParentID: req.GetString("category_id", ""),
				Topic:    req.GetString("topic", ""),
				NSFW:     req.GetBool("nsfw", false),
			}
			if rl := req.GetFloat("rate_limit", 0); rl > 0 {
				data.RateLimitPerUser = int(rl)
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	// create_voice_channel
	s.AddTool(
		mcp.NewTool("create_voice_channel",
			mcp.WithDescription("Create a new voice channel"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("category_id", mcp.Description("Parent category ID")),
			mcp.WithNumber("bitrate", mcp.Description("Bitrate in bits (8000-384000)")),
			mcp.WithNumber("user_limit", mcp.Description("Max users (0 = unlimited, max 99)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error())
			}
			data := discordgo.GuildChannelCreateData{
				Name:     name,
				Type:     discordgo.ChannelTypeGuildVoice,
				ParentID: req.GetString("category_id", ""),
				Bitrate:  int(req.GetFloat("bitrate", 64000)),
			}
			if ul := req.GetFloat("user_limit", 0); ul > 0 {
				data.UserLimit = int(ul)
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, data)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	// create_category
	s.AddTool(
		mcp.NewTool("create_category",
			mcp.WithDescription("Create a new channel category"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Category name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error())
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
				Name: name,
				Type: discordgo.ChannelTypeGuildCategory,
			})
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	// edit_channel
	s.AddTool(
		mcp.NewTool("edit_channel",
			mcp.WithDescription("Edit an existing channel's properties"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID to edit")),
			mcp.WithString("name", mcp.Description("New channel name")),
			mcp.WithString("topic", mcp.Description("New channel topic")),
			mcp.WithBoolean("nsfw", mcp.Description("Whether NSFW")),
			mcp.WithNumber("rate_limit", mcp.Description("Slowmode in seconds")),
			mcp.WithString("category_id", mcp.Description("Move to category (empty string to remove from category)")),
			mcp.WithNumber("position", mcp.Description("Channel position")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
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
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	// delete_channel
	s.AddTool(
		mcp.NewTool("delete_channel",
			mcp.WithDescription("Delete a channel (irreversible)"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID to delete")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
			}
			_, err = bot.ChannelDelete(channelID)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("channel deleted"), nil
		},
	)

	// set_channel_permissions
	s.AddTool(
		mcp.NewTool("set_channel_permissions",
			mcp.WithDescription("Set permission overrides for a role or user on a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID to set permissions on")),
			mcp.WithString("target_id", mcp.Required(), mcp.Description("Role ID or User ID to set permissions for")),
			mcp.WithString("target_type", mcp.Required(), mcp.Description("Type of target: 'role' or 'member'")),
			mcp.WithString("allow", mcp.Description("Allowed permissions as bitfield string (e.g. '3072' for VIEW_CHANNEL+SEND_MESSAGES)")),
			mcp.WithString("deny", mcp.Description("Denied permissions as bitfield string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
			}
			targetID, err := req.RequireString("target_id")
			if err != nil {
				return toolError(err.Error())
			}
			targetType, err := req.RequireString("target_type")
			if err != nil {
				return toolError(err.Error())
			}

			var dType discordgo.PermissionOverwriteType
			if targetType == "role" {
				dType = discordgo.PermissionOverwriteTypeRole
			} else {
				dType = discordgo.PermissionOverwriteTypeMember
			}

			allow := int64(0)
			if v := req.GetString("allow", ""); v != "" {
				allow, _ = strconv.ParseInt(v, 10, 64)
			}
			deny := int64(0)
			if v := req.GetString("deny", ""); v != "" {
				deny, _ = strconv.ParseInt(v, 10, 64)
			}

			err = bot.ChannelPermissionSet(channelID, targetID, dType, allow, deny)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]interface{}{
				"channel_id":  channelID,
				"target_id":   targetID,
				"target_type": targetType,
				"allow":       allow,
				"allow_names": describePermissionBits(allow),
				"deny":        deny,
				"deny_names":  describePermissionBits(deny),
			})
		},
	)

	// sync_category_permissions
	s.AddTool(
		mcp.NewTool("sync_category_permissions",
			mcp.WithDescription("Set permission overrides on a category AND sync to all its child channels"),
			mcp.WithString("category_id", mcp.Required(), mcp.Description("Category channel ID")),
			mcp.WithString("target_id", mcp.Required(), mcp.Description("Role ID or User ID")),
			mcp.WithString("target_type", mcp.Required(), mcp.Description("Type: 'role' or 'member'")),
			mcp.WithString("allow", mcp.Description("Allowed permissions as bitfield string")),
			mcp.WithString("deny", mcp.Description("Denied permissions as bitfield string")),
			mcp.WithBoolean("force", mcp.Description("If true, overwrite existing custom overrides on child channels")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			categoryID, err := req.RequireString("category_id")
			if err != nil {
				return toolError(err.Error())
			}
			targetID, err := req.RequireString("target_id")
			if err != nil {
				return toolError(err.Error())
			}
			targetType, err := req.RequireString("target_type")
			if err != nil {
				return toolError(err.Error())
			}

			cat, err := bot.Channel(categoryID)
			if err != nil {
				return toolError("category not found: " + err.Error())
			}
			if cat.Type != discordgo.ChannelTypeGuildCategory {
				return toolError("channel is not a category")
			}

			var dType discordgo.PermissionOverwriteType
			if targetType == "role" {
				dType = discordgo.PermissionOverwriteTypeRole
			} else {
				dType = discordgo.PermissionOverwriteTypeMember
			}

			allow := int64(0)
			if v := req.GetString("allow", ""); v != "" {
				allow, _ = strconv.ParseInt(v, 10, 64)
			}
			deny := int64(0)
			if v := req.GetString("deny", ""); v != "" {
				deny, _ = strconv.ParseInt(v, 10, 64)
			}
			force := req.GetBool("force", false)

			err = bot.ChannelPermissionSet(categoryID, targetID, dType, allow, deny)
			if err != nil {
				return toolError("failed to set category permissions: " + err.Error())
			}

			channels, err := bot.GuildChannels(guildID)
			if err != nil {
				return toolError("failed to list channels: " + err.Error())
			}

			synced := []string{}
			skipped := []string{}
			failed := []string{}

			for _, ch := range channels {
				if ch.ParentID != categoryID {
					continue
				}
				if !force {
					hasCustomOverride := false
					for _, perm := range ch.PermissionOverwrites {
						if perm.ID == targetID {
							hasCustomOverride = true
							break
						}
					}
					if hasCustomOverride {
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

			return resultJSON(map[string]interface{}{
				"category_name": cat.Name,
				"target_type":   targetType,
				"allow_names":   describePermissionBits(allow),
				"deny_names":    describePermissionBits(deny),
				"synced":        synced,
				"skipped":       skipped,
				"failed":        failed,
			})
		},
	)
}
