package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAdvancedChannelTools(s *server.MCPServer, d *Discord) {
	bot := d.Session
	guildID := d.GuildID

	s.AddTool(
		mcp.NewTool("create_announcement_channel",
			mcp.WithDescription("Create an announcement (news) channel. Messages can be crossposted to following servers."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("category_id", mcp.Description("Parent category ID")),
			mcp.WithString("topic", mcp.Description("Channel topic")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error())
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
				Name:     name,
				Type:     discordgo.ChannelTypeGuildNews,
				ParentID: req.GetString("category_id", ""),
				Topic:    req.GetString("topic", ""),
			})
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("create_stage_channel",
			mcp.WithDescription("Create a stage channel for audio events and presentations"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("category_id", mcp.Description("Parent category ID")),
			mcp.WithString("topic", mcp.Description("Stage topic")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error())
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
				Name:     name,
				Type:     discordgo.ChannelTypeGuildStageVoice,
				ParentID: req.GetString("category_id", ""),
				Topic:    req.GetString("topic", ""),
			})
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("create_forum_channel",
			mcp.WithDescription("Create a forum channel where members can create organized discussion posts"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("category_id", mcp.Description("Parent category ID")),
			mcp.WithString("topic", mcp.Description("Forum guidelines/topic")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error())
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
				Name:     name,
				Type:     discordgo.ChannelTypeGuildForum,
				ParentID: req.GetString("category_id", ""),
				Topic:    req.GetString("topic", ""),
			})
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("create_forum_post",
			mcp.WithDescription("Create a new post (thread) in a forum channel with an initial message"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Forum channel ID")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Post title")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Initial message content")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
			}
			title, err := req.RequireString("title")
			if err != nil {
				return toolError(err.Error())
			}
			content, err := req.RequireString("content")
			if err != nil {
				return toolError(err.Error())
			}
			thread, err := bot.ForumThreadStartComplex(channelID, &discordgo.ThreadStart{
				Name:                title,
				AutoArchiveDuration: 1440,
			}, &discordgo.MessageSend{
				Content: content,
			})
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": thread.ID, "name": thread.Name})
		},
	)
}
