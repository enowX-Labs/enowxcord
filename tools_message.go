package main

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMessageTools(s *server.MCPServer, d *Discord) {
	bot := d.Session

	s.AddTool(
		mcp.NewTool("send_message",
			mcp.WithDescription("Send a message to a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID to send message to")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Message content (max 2000 chars)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
			}
			content, err := req.RequireString("content")
			if err != nil {
				return toolError(err.Error())
			}
			msg, err := bot.ChannelMessageSend(channelID, content)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"message_id": msg.ID})
		},
	)

	s.AddTool(
		mcp.NewTool("send_embed",
			mcp.WithDescription("Send a rich embed message to a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("title", mcp.Description("Embed title")),
			mcp.WithString("description", mcp.Description("Embed description")),
			mcp.WithNumber("color", mcp.Description("Embed color as decimal integer")),
			mcp.WithString("footer", mcp.Description("Footer text")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
			}
			embed := &discordgo.MessageEmbed{
				Title:       req.GetString("title", ""),
				Description: req.GetString("description", ""),
				Color:       int(req.GetFloat("color", 3447003)),
			}
			if footer := req.GetString("footer", ""); footer != "" {
				embed.Footer = &discordgo.MessageEmbedFooter{Text: footer}
			}
			msg, err := bot.ChannelMessageSendEmbed(channelID, embed)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"message_id": msg.ID})
		},
	)

	s.AddTool(
		mcp.NewTool("bulk_delete_messages",
			mcp.WithDescription("Delete multiple messages from a channel (2-100 messages, max 14 days old)"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithNumber("count", mcp.Required(), mcp.Description("Number of recent messages to delete (2-100)")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
			}
			count, err := req.RequireFloat("count")
			if err != nil {
				return toolError(err.Error())
			}
			c := int(count)
			if c < 2 {
				return toolError("count must be at least 2")
			}
			if c > 100 {
				c = 100
			}
			messages, err := bot.ChannelMessages(channelID, c, "", "", "")
			if err != nil {
				return toolError(err.Error())
			}
			ids := make([]string, 0, len(messages))
			for _, m := range messages {
				ids = append(ids, m.ID)
			}
			if len(ids) < 2 {
				return toolError("not enough messages to delete")
			}
			err = bot.ChannelMessagesBulkDelete(channelID, ids)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText(fmt.Sprintf("deleted %d messages", len(ids))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("pin_message",
			mcp.WithDescription("Pin a message in a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID to pin")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
			}
			messageID, err := req.RequireString("message_id")
			if err != nil {
				return toolError(err.Error())
			}
			err = bot.ChannelMessagePin(channelID, messageID)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("message pinned"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_thread",
			mcp.WithDescription("Create a new thread in a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Parent channel ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Thread name")),
			mcp.WithNumber("auto_archive_duration", mcp.Description("Auto-archive after minutes of inactivity (60, 1440, 4320, 10080)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return toolError(err.Error())
			}
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error())
			}
			archiveDuration := int(req.GetFloat("auto_archive_duration", 1440))
			thread, err := bot.ThreadStart(channelID, name, discordgo.ChannelTypeGuildPublicThread, archiveDuration)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{"id": thread.ID, "name": thread.Name})
		},
	)
}
