package member

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

func Register(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("list_members",
			mcp.WithDescription("List server members (up to 100)"),
			mcp.WithNumber("limit", mcp.Description("Max members to return (1-100, default 50)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			limit := int(req.GetFloat("limit", 50))
			if limit > 100 {
				limit = 100
			}
			if limit < 1 {
				limit = 50
			}
			members, err := bot.GuildMembers(guildID, "", limit)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				UserID   string   `json:"user_id"`
				Username string   `json:"username"`
				Nick     string   `json:"nick,omitempty"`
				Roles    []string `json:"roles"`
				JoinedAt string   `json:"joined_at"`
			}
			result := make([]entry, 0, len(members))
			for _, m := range members {
				result = append(result, entry{
					UserID: m.User.ID, Username: m.User.Username,
					Nick: m.Nick, Roles: m.Roles,
					JoinedAt: m.JoinedAt.Format(time.RFC3339),
				})
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("get_member",
			mcp.WithDescription("Get detailed information about a specific member"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			m, err := bot.GuildMember(guildID, userID)
			if err != nil {
				return tools.Error(err.Error())
			}
			result := map[string]interface{}{
				"nick": m.Nick, "roles": m.Roles,
				"joined_at": m.JoinedAt.Format(time.RFC3339),
			}
			if m.User != nil {
				result["user_id"] = m.User.ID
				result["username"] = m.User.Username
				result["avatar"] = m.User.Avatar
				result["bot"] = m.User.Bot
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("edit_member",
			mcp.WithDescription("Edit a member's properties (nickname, roles, mute, deaf)"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID")),
			mcp.WithString("nick", mcp.Description("New nickname (empty to reset)")),
			mcp.WithArray("roles", mcp.Description("Array of role IDs to set (replaces all roles)"), mcp.Items(map[string]any{"type": "string"})),
			mcp.WithBoolean("mute", mcp.Description("Server mute")),
			mcp.WithBoolean("deaf", mcp.Description("Server deafen")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			params := &discordgo.GuildMemberParams{}
			args := req.GetArguments()
			if v, ok := args["nick"]; ok {
				if s, ok := v.(string); ok {
					params.Nick = s
				}
			}
			if v, ok := args["roles"]; ok {
				if arr, ok := v.([]interface{}); ok {
					roles := make([]string, 0, len(arr))
					for _, r := range arr {
						if s, ok := r.(string); ok {
							roles = append(roles, s)
						}
					}
					params.Roles = &roles
				}
			}
			if _, ok := args["mute"]; ok {
				b := req.GetBool("mute", false)
				params.Mute = &b
			}
			if _, ok := args["deaf"]; ok {
				b := req.GetBool("deaf", false)
				params.Deaf = &b
			}
			if _, err = bot.GuildMemberEditComplex(guildID, userID, params); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("member updated"), nil
		},
	)

	s.AddTool(mcp.NewTool("add_role_to_member", mcp.WithDescription("Add a role to a member"), mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID")), mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID to add"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, _ := req.RequireString("user_id")
			roleID, _ := req.RequireString("role_id")
			if userID == "" || roleID == "" {
				return tools.Error("user_id and role_id are required")
			}
			if err := bot.GuildMemberRoleAdd(guildID, userID, roleID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("role added"), nil
		},
	)

	s.AddTool(mcp.NewTool("remove_role_from_member", mcp.WithDescription("Remove a role from a member"), mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID")), mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID to remove"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, _ := req.RequireString("user_id")
			roleID, _ := req.RequireString("role_id")
			if userID == "" || roleID == "" {
				return tools.Error("user_id and role_id are required")
			}
			if err := bot.GuildMemberRoleRemove(guildID, userID, roleID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("role removed"), nil
		},
	)

	s.AddTool(mcp.NewTool("kick_member", mcp.WithDescription("Kick a member from the server"), mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to kick")), mcp.WithString("reason", mcp.Description("Reason for kick")), mcp.WithDestructiveHintAnnotation(true)),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.GuildMemberDeleteWithReason(guildID, userID, req.GetString("reason", "")); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("member kicked"), nil
		},
	)

	s.AddTool(mcp.NewTool("ban_member", mcp.WithDescription("Ban a member from the server (permanent until unbanned)"), mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to ban")), mcp.WithString("reason", mcp.Description("Reason for ban")), mcp.WithNumber("delete_days", mcp.Description("Days of messages to delete (0-7)")), mcp.WithDestructiveHintAnnotation(true)),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.GuildBanCreateWithReason(guildID, userID, req.GetString("reason", ""), int(req.GetFloat("delete_days", 0))); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("member banned"), nil
		},
	)

	s.AddTool(mcp.NewTool("unban_member", mcp.WithDescription("Unban a previously banned user"), mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to unban"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.GuildBanDelete(guildID, userID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("member unbanned"), nil
		},
	)

	s.AddTool(mcp.NewTool("timeout_member", mcp.WithDescription("Timeout (mute) a member for a duration"), mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to timeout")), mcp.WithNumber("duration_seconds", mcp.Required(), mcp.Description("Timeout duration in seconds (max 2419200 = 28 days)")), mcp.WithString("reason", mcp.Description("Reason for timeout"))),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			userID, err := req.RequireString("user_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			duration, err := req.RequireFloat("duration_seconds")
			if err != nil {
				return tools.Error(err.Error())
			}
			until := time.Now().Add(time.Duration(duration) * time.Second)
			if err = bot.GuildMemberTimeout(guildID, userID, &until); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText(fmt.Sprintf("timed out until %s", until.Format(time.RFC3339))), nil
		},
	)

	s.AddTool(mcp.NewTool("list_bans", mcp.WithDescription("List all banned users")),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			bans, err := bot.GuildBans(guildID, 100, "", "")
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				Reason   string `json:"reason,omitempty"`
			}
			result := make([]entry, 0, len(bans))
			for _, b := range bans {
				result = append(result, entry{UserID: b.User.ID, Username: b.User.Username, Reason: b.Reason})
			}
			return tools.JSON(result)
		},
	)
}
