package thread

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

func Register(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("list_active_threads",
			mcp.WithDescription("List all active (non-archived) threads in the server"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			threads, err := bot.GuildThreadsActive(guildID)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				ParentID string `json:"parent_id"`
				Archived bool   `json:"archived"`
				Locked   bool   `json:"locked"`
			}
			result := make([]entry, 0, len(threads.Threads))
			for _, t := range threads.Threads {
				e := entry{ID: t.ID, Name: t.Name, ParentID: t.ParentID}
				if t.ThreadMetadata != nil {
					e.Archived = t.ThreadMetadata.Archived
					e.Locked = t.ThreadMetadata.Locked
				}
				result = append(result, e)
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("create_thread_from_message",
			mcp.WithDescription("Create a thread attached to an existing message"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID containing the message")),
			mcp.WithString("message_id", mcp.Required(), mcp.Description("Message ID to start the thread from")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Thread name")),
			mcp.WithNumber("auto_archive_duration", mcp.Description("Auto-archive after minutes (60, 1440, 4320, 10080; default 1440)")),
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
			name, err := req.RequireString("name")
			if err != nil {
				return tools.Error(err.Error())
			}
			t, err := bot.MessageThreadStart(channelID, messageID, name, int(req.GetFloat("auto_archive_duration", 1440)))
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": t.ID, "name": t.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("set_thread_state",
			mcp.WithDescription("Archive/unarchive or lock/unlock a thread"),
			mcp.WithString("thread_id", mcp.Required(), mcp.Description("Thread (channel) ID")),
			mcp.WithBoolean("archived", mcp.Description("Set archived state")),
			mcp.WithBoolean("locked", mcp.Description("Set locked state (only moderators can unarchive)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			threadID, err := req.RequireString("thread_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			edit := &discordgo.ChannelEdit{}
			args := req.GetArguments()
			if _, ok := args["archived"]; ok {
				b := req.GetBool("archived", false)
				edit.Archived = &b
			}
			if _, ok := args["locked"]; ok {
				b := req.GetBool("locked", false)
				edit.Locked = &b
			}
			t, err := bot.ChannelEditComplex(threadID, edit)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"id": t.ID, "name": t.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("add_thread_member",
			mcp.WithDescription("Add a member to a thread"),
			mcp.WithString("thread_id", mcp.Required(), mcp.Description("Thread ID")),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to add")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			threadID, err := req.RequireString("thread_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.ThreadMemberAdd(threadID, userID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("member added to thread"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("remove_thread_member",
			mcp.WithDescription("Remove a member from a thread"),
			mcp.WithString("thread_id", mcp.Required(), mcp.Description("Thread ID")),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to remove")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			threadID, err := req.RequireString("thread_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.ThreadMemberRemove(threadID, userID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("member removed from thread"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_thread_members",
			mcp.WithDescription("List members of a thread"),
			mcp.WithString("thread_id", mcp.Required(), mcp.Description("Thread ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, _, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			threadID, err := req.RequireString("thread_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			members, err := bot.ThreadMembers(threadID, 100, false, "")
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				UserID   string `json:"user_id"`
				JoinedAt string `json:"joined_at"`
			}
			result := make([]entry, 0, len(members))
			for _, m := range members {
				result = append(result, entry{UserID: m.UserID, JoinedAt: m.JoinTimestamp.Format(time.RFC3339)})
			}
			return tools.JSON(result)
		},
	)
}
