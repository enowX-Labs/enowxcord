package channel

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

func registerAdvanced(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("create_announcement_channel",
			mcp.WithDescription("Create an announcement (news) channel"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("category_id", mcp.Description("Parent category ID")),
			mcp.WithString("topic", mcp.Description("Channel topic")),
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
				Name: name, Type: discordgo.ChannelTypeGuildNews,
				ParentID: req.GetString("category_id", ""),
				Topic:    req.GetString("topic", ""),
			})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": ch.ID, "name": ch.Name})
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
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			name, err := req.RequireString("name")
			if err != nil {
				return tools.Error(err.Error())
			}
			ch, err := bot.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
				Name: name, Type: discordgo.ChannelTypeGuildStageVoice,
				ParentID: req.GetString("category_id", ""),
				Topic:    req.GetString("topic", ""),
			})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": ch.ID, "name": ch.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("create_forum_channel",
			mcp.WithDescription("Create a forum channel for organized discussion posts"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Channel name")),
			mcp.WithString("category_id", mcp.Description("Parent category ID")),
			mcp.WithString("topic", mcp.Description("Forum guidelines/topic")),
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
				Name: name, Type: discordgo.ChannelTypeGuildForum,
				ParentID: req.GetString("category_id", ""),
				Topic:    req.GetString("topic", ""),
			})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": ch.ID, "name": ch.Name})
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
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			title, err := req.RequireString("title")
			if err != nil {
				return tools.Error(err.Error())
			}
			content, err := req.RequireString("content")
			if err != nil {
				return tools.Error(err.Error())
			}
			thread, err := bot.ForumThreadStartComplex(channelID, &discordgo.ThreadStart{
				Name: title, AutoArchiveDuration: 1440,
			}, &discordgo.MessageSend{Content: content})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": thread.ID, "name": thread.Name})
		},
	)
}
