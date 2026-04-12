package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMemberTools(s *server.MCPServer, d *Discord) {
	bot := d.Session
	guildID := d.GuildID

	s.AddTool(
		mcp.NewTool("list_members",
			mcp.WithDescription("List server members (up to 100)"),
			mcp.WithNumber("limit", mcp.Description("Max members to return (1-100, default 50)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			limit := int(req.GetFloat("limit", 50))
			if limit > 100 {
				limit = 100
			}
			if limit < 1 {
				limit = 50
			}
			members, err := bot.GuildMembers(guildID, "", limit)
			if err != nil {
				return toolError(err.Error())
			}
			type m struct {
				UserID   string   `json:"user_id"`
				Username string   `json:"username"`
				Nick     string   `json:"nick,omitempty"`
				Roles    []string `json:"roles"`
				JoinedAt string   `json:"joined_at"`
			}
			result := make([]m, 0, len(members))
			for _, member := range members {
				result = append(result, m{
					UserID:   member.User.ID,
					Username: member.User.Username,
					Nick:     member.Nick,
					Roles:    member.Roles,
					JoinedAt: member.JoinedAt.Format(time.RFC3339),
				})
			}
			return resultJSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("get_member",
			mcp.WithDescription("Get detailed information about a specific member"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID, err := req.RequireString("user_id")
			if err != nil {
				return toolError(err.Error())
			}
			member, err := bot.GuildMember(guildID, userID)
			if err != nil {
				return toolError(err.Error())
			}
			result := map[string]interface{}{
				"nick":      member.Nick,
				"roles":     member.Roles,
				"joined_at": member.JoinedAt.Format(time.RFC3339),
			}
			if member.User != nil {
				result["user_id"] = member.User.ID
				result["username"] = member.User.Username
				result["avatar"] = member.User.Avatar
				result["bot"] = member.User.Bot
			}
			return resultJSON(result)
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
			userID, err := req.RequireString("user_id")
			if err != nil {
				return toolError(err.Error())
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
			_, err = bot.GuildMemberEditComplex(guildID, userID, params)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("member updated"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("add_role_to_member",
			mcp.WithDescription("Add a role to a member"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID")),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID to add")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID, err := req.RequireString("user_id")
			if err != nil {
				return toolError(err.Error())
			}
			roleID, err := req.RequireString("role_id")
			if err != nil {
				return toolError(err.Error())
			}
			err = bot.GuildMemberRoleAdd(guildID, userID, roleID)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("role added"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("remove_role_from_member",
			mcp.WithDescription("Remove a role from a member"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID")),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID to remove")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID, err := req.RequireString("user_id")
			if err != nil {
				return toolError(err.Error())
			}
			roleID, err := req.RequireString("role_id")
			if err != nil {
				return toolError(err.Error())
			}
			err = bot.GuildMemberRoleRemove(guildID, userID, roleID)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("role removed"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("kick_member",
			mcp.WithDescription("Kick a member from the server (they can rejoin with invite)"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to kick")),
			mcp.WithString("reason", mcp.Description("Reason for kick")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID, err := req.RequireString("user_id")
			if err != nil {
				return toolError(err.Error())
			}
			reason := req.GetString("reason", "")
			err = bot.GuildMemberDeleteWithReason(guildID, userID, reason)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("member kicked"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("ban_member",
			mcp.WithDescription("Ban a member from the server (permanent until unbanned)"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to ban")),
			mcp.WithString("reason", mcp.Description("Reason for ban")),
			mcp.WithNumber("delete_days", mcp.Description("Days of messages to delete (0-7)")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID, err := req.RequireString("user_id")
			if err != nil {
				return toolError(err.Error())
			}
			reason := req.GetString("reason", "")
			deleteDays := int(req.GetFloat("delete_days", 0))
			err = bot.GuildBanCreateWithReason(guildID, userID, reason, deleteDays)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("member banned"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("unban_member",
			mcp.WithDescription("Unban a previously banned user"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to unban")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID, err := req.RequireString("user_id")
			if err != nil {
				return toolError(err.Error())
			}
			err = bot.GuildBanDelete(guildID, userID)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("member unbanned"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("timeout_member",
			mcp.WithDescription("Timeout (mute) a member for a duration"),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("User ID to timeout")),
			mcp.WithNumber("duration_seconds", mcp.Required(), mcp.Description("Timeout duration in seconds (max 2419200 = 28 days)")),
			mcp.WithString("reason", mcp.Description("Reason for timeout")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userID, err := req.RequireString("user_id")
			if err != nil {
				return toolError(err.Error())
			}
			duration, err := req.RequireFloat("duration_seconds")
			if err != nil {
				return toolError(err.Error())
			}
			until := time.Now().Add(time.Duration(duration) * time.Second)
			err = bot.GuildMemberTimeout(guildID, userID, &until)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText(fmt.Sprintf("timed out until %s", until.Format(time.RFC3339))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_bans",
			mcp.WithDescription("List all banned users"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bans, err := bot.GuildBans(guildID, 100, "", "")
			if err != nil {
				return toolError(err.Error())
			}
			type b struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				Reason   string `json:"reason,omitempty"`
			}
			result := make([]b, 0, len(bans))
			for _, ban := range bans {
				result = append(result, b{UserID: ban.User.ID, Username: ban.User.Username, Reason: ban.Reason})
			}
			return resultJSON(result)
		},
	)
}
