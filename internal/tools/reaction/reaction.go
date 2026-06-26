package reaction

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

// emojiDesc documents the emoji argument format shared across reaction tools.
const emojiDesc = "Emoji: a unicode emoji (e.g. 👍) or a custom emoji as 'name:id'"

func Register(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("add_reaction",
			mcp.WithDescription("Add a reaction to a message"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID")),
			mcp.WithString("emoji", mcp.Required(), mcp.Description(emojiDesc)),
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
			emoji, err := req.RequireString("emoji")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.MessageReactionAdd(channelID, messageID, emoji); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("reaction added"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("remove_reaction",
			mcp.WithDescription("Remove a reaction from a message. Removes the bot's own reaction by default, or a specific user's if user_id is given."),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID")),
			mcp.WithString("emoji", mcp.Required(), mcp.Description(emojiDesc)),
			mcp.WithString("user_id", mcp.Description("User ID whose reaction to remove (default: the bot, '@me')")),
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
			emoji, err := req.RequireString("emoji")
			if err != nil {
				return tools.Error(err.Error())
			}
			userID := req.GetString("user_id", "@me")
			if err = bot.MessageReactionRemove(channelID, messageID, emoji, userID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("reaction removed"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("remove_all_reactions",
			mcp.WithDescription("Remove all reactions from a message, or all of one emoji if 'emoji' is given"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID")),
			mcp.WithString("emoji", mcp.Description("If set, only remove reactions of this emoji ("+emojiDesc+")")),
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
			if emoji := req.GetString("emoji", ""); emoji != "" {
				if err = bot.MessageReactionsRemoveEmoji(channelID, messageID, emoji); err != nil {
					return tools.Error(err.Error())
				}
				return mcp.NewToolResultText("reactions of emoji removed"), nil
			}
			if err = bot.MessageReactionsRemoveAll(channelID, messageID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("all reactions removed"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_reactions",
			mcp.WithDescription("List users who reacted to a message with a specific emoji"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID")),
			mcp.WithString("emoji", mcp.Required(), mcp.Description(emojiDesc)),
			mcp.WithNumber("limit", mcp.Description("Max users to return (1-100, default 100)")),
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
			emoji, err := req.RequireString("emoji")
			if err != nil {
				return tools.Error(err.Error())
			}
			limit := int(req.GetFloat("limit", 100))
			if limit < 1 || limit > 100 {
				limit = 100
			}
			users, err := bot.MessageReactions(channelID, messageID, emoji, limit, "", "")
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
			}
			result := make([]entry, 0, len(users))
			for _, u := range users {
				result = append(result, entry{UserID: u.ID, Username: u.Username})
			}
			return tools.JSON(result)
		},
	)
}
