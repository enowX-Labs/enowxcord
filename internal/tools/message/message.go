package message

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

// buildEmbed constructs a MessageEmbed from request arguments. It supports
// title, description, color, url, footer, author, image, thumbnail, timestamp
// and a JSON-encoded array of fields ([{"name","value","inline"}]).
func buildEmbed(req mcp.CallToolRequest) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       req.GetString("title", ""),
		Description: req.GetString("description", ""),
		URL:         req.GetString("url", ""),
		Color:       int(req.GetFloat("color", 3447003)),
	}
	if footer := req.GetString("footer", ""); footer != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: footer}
	}
	if author := req.GetString("author", ""); author != "" {
		embed.Author = &discordgo.MessageEmbedAuthor{Name: author, IconURL: req.GetString("author_icon", "")}
	}
	if img := req.GetString("image", ""); img != "" {
		embed.Image = &discordgo.MessageEmbedImage{URL: img}
	}
	if thumb := req.GetString("thumbnail", ""); thumb != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: thumb}
	}
	if req.GetBool("timestamp", false) {
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}
	if fields, ok := req.GetArguments()["fields"].([]any); ok {
		for _, f := range fields {
			m, ok := f.(map[string]any)
			if !ok {
				continue
			}
			name, _ := m["name"].(string)
			value, _ := m["value"].(string)
			inline, _ := m["inline"].(bool)
			if name == "" {
				continue
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: name, Value: value, Inline: inline})
		}
	}
	return embed
}

func Register(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("send_message", mcp.WithDescription("Send a message to a channel"), mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID to send message to")), mcp.WithString("content", mcp.Required(), mcp.Description("Message content (max 2000 chars)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			content, err := req.RequireString("content")
			if err != nil {
				return tools.Error(err.Error())
			}
			msg, err := bot.ChannelMessageSend(channelID, content)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"message_id": msg.ID})
		},
	)

	s.AddTool(
		mcp.NewTool("send_embed",
			mcp.WithDescription("Send a rich embed message to a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("title", mcp.Description("Embed title")),
			mcp.WithString("description", mcp.Description("Embed description")),
			mcp.WithNumber("color", mcp.Description("Embed color as decimal integer (default 3447003)")),
			mcp.WithString("url", mcp.Description("URL the title links to")),
			mcp.WithString("footer", mcp.Description("Footer text")),
			mcp.WithString("author", mcp.Description("Author name shown above the title")),
			mcp.WithString("author_icon", mcp.Description("Author icon URL")),
			mcp.WithString("image", mcp.Description("Large image URL")),
			mcp.WithString("thumbnail", mcp.Description("Thumbnail image URL")),
			mcp.WithBoolean("timestamp", mcp.Description("Set the embed timestamp to now")),
			mcp.WithArray("fields", mcp.Description("Embed fields: array of {name, value, inline}"), mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":   map[string]any{"type": "string"},
					"value":  map[string]any{"type": "string"},
					"inline": map[string]any{"type": "boolean"},
				},
			})),
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
			msg, err := bot.ChannelMessageSendEmbed(channelID, buildEmbed(req))
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"message_id": msg.ID})
		},
	)

	s.AddTool(
		mcp.NewTool("bulk_delete_messages", mcp.WithDescription("Delete multiple messages from a channel (2-100 messages, max 14 days old)"), mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")), mcp.WithNumber("count", mcp.Required(), mcp.Description("Number of recent messages to delete (2-100)")), mcp.WithDestructiveHintAnnotation(true)),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			count, err := req.RequireFloat("count")
			if err != nil {
				return tools.Error(err.Error())
			}
			c := int(count)
			if c < 2 {
				return tools.Error("count must be at least 2")
			}
			if c > 100 {
				c = 100
			}
			messages, err := bot.ChannelMessages(channelID, c, "", "", "")
			if err != nil {
				return tools.Error(err.Error())
			}
			ids := make([]string, 0, len(messages))
			for _, m := range messages {
				ids = append(ids, m.ID)
			}
			if len(ids) < 2 {
				return tools.Error("not enough messages to delete")
			}
			if err = bot.ChannelMessagesBulkDelete(channelID, ids); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText(fmt.Sprintf("deleted %d messages", len(ids))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("pin_message", mcp.WithDescription("Pin a message in a channel"), mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")), mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID to pin"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			messageID, err := req.RequireString("message_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.ChannelMessagePin(channelID, messageID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("message pinned"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_thread", mcp.WithDescription("Create a new thread in a channel"), mcp.WithString("channel_id", mcp.Required(), mcp.Description("Parent channel ID")), mcp.WithString("name", mcp.Required(), mcp.Description("Thread name")), mcp.WithNumber("auto_archive_duration", mcp.Description("Auto-archive after minutes of inactivity (60, 1440, 4320, 10080)"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			name, err := req.RequireString("name")
			if err != nil {
				return tools.Error(err.Error())
			}
			thread, err := bot.ThreadStart(channelID, name, discordgo.ChannelTypeGuildPublicThread, int(req.GetFloat("auto_archive_duration", 1440)))
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": thread.ID, "name": thread.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("get_messages",
			mcp.WithDescription("Read recent messages from a channel (newest first)"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithNumber("limit", mcp.Description("Number of messages to fetch (1-100, default 50)")),
			mcp.WithString("before", mcp.Description("Return messages before this message ID")),
			mcp.WithString("after", mcp.Description("Return messages after this message ID")),
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
			limit := int(req.GetFloat("limit", 50))
			if limit < 1 {
				limit = 50
			}
			if limit > 100 {
				limit = 100
			}
			messages, err := bot.ChannelMessages(channelID, limit, req.GetString("before", ""), req.GetString("after", ""), "")
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(summarizeMessages(messages))
		},
	)

	s.AddTool(
		mcp.NewTool("get_message",
			mcp.WithDescription("Get a single message by ID"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID")),
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
			messageID, err := req.RequireString("message_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			msg, err := bot.ChannelMessage(channelID, messageID)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(summarizeMessage(msg))
		},
	)

	s.AddTool(
		mcp.NewTool("edit_message",
			mcp.WithDescription("Edit the text content of a message the bot sent"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID to edit")),
			mcp.WithString("content", mcp.Required(), mcp.Description("New message content")),
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
			messageID, err := req.RequireString("message_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			content, err := req.RequireString("content")
			if err != nil {
				return tools.Error(err.Error())
			}
			if _, err = bot.ChannelMessageEdit(channelID, messageID, content); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("message edited"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("delete_message",
			mcp.WithDescription("Delete a single message"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID to delete")),
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
			messageID, err := req.RequireString("message_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.ChannelMessageDelete(channelID, messageID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("message deleted"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("reply_message",
			mcp.WithDescription("Send a message as a reply to another message"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID to reply to")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Reply content")),
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
			messageID, err := req.RequireString("message_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			content, err := req.RequireString("content")
			if err != nil {
				return tools.Error(err.Error())
			}
			msg, err := bot.ChannelMessageSendReply(channelID, content, &discordgo.MessageReference{MessageID: messageID, ChannelID: channelID})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"message_id": msg.ID})
		},
	)

	s.AddTool(
		mcp.NewTool("unpin_message",
			mcp.WithDescription("Unpin a message in a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID to unpin")),
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
			messageID, err := req.RequireString("message_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.ChannelMessageUnpin(channelID, messageID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("message unpinned"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_pinned_messages",
			mcp.WithDescription("List all pinned messages in a channel"),
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
			messages, err := bot.ChannelMessagesPinned(channelID)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(summarizeMessages(messages))
		},
	)

	s.AddTool(
		mcp.NewTool("send_dm",
			mcp.WithDescription("Send a direct message to a user"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to DM")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Message content")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, errResult := tools.BotFromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			content, err := req.RequireString("content")
			if err != nil {
				return tools.Error(err.Error())
			}
			ch, err := bot.UserChannelCreate(userID)
			if err != nil {
				return tools.Errorf("failed to open DM channel: %v", err)
			}
			msg, err := bot.ChannelMessageSend(ch.ID, content)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"channel_id": ch.ID, "message_id": msg.ID})
		},
	)

	s.AddTool(
		mcp.NewTool("crosspost_message",
			mcp.WithDescription("Publish (crosspost) a message from an announcement channel to following servers"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Announcement channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID to publish")),
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
			messageID, err := req.RequireString("message_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if _, err = bot.ChannelMessageCrosspost(channelID, messageID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("message published"), nil
		},
	)
}

type messageSummary struct {
	ID          string   `json:"id"`
	Author      string   `json:"author"`
	AuthorID    string   `json:"author_id"`
	Content     string   `json:"content"`
	Timestamp   string   `json:"timestamp"`
	Pinned      bool     `json:"pinned,omitempty"`
	EmbedCount  int      `json:"embed_count,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

func summarizeMessage(m *discordgo.Message) messageSummary {
	s := messageSummary{
		ID: m.ID, Content: m.Content, Pinned: m.Pinned,
		Timestamp: m.Timestamp.Format(time.RFC3339), EmbedCount: len(m.Embeds),
	}
	if m.Author != nil {
		s.Author = m.Author.Username
		s.AuthorID = m.Author.ID
	}
	for _, a := range m.Attachments {
		s.Attachments = append(s.Attachments, a.URL)
	}
	return s
}

func summarizeMessages(messages []*discordgo.Message) []messageSummary {
	result := make([]messageSummary, 0, len(messages))
	for _, m := range messages {
		result = append(result, summarizeMessage(m))
	}
	return result
}
