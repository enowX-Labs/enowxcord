package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRoleTools(s *server.MCPServer, d *Discord) {
	bot := d.Session
	guildID := d.GuildID

	s.AddTool(
		mcp.NewTool("list_roles",
			mcp.WithDescription("List all roles in the server with their permissions, colors, and positions"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			roles, err := bot.GuildRoles(guildID)
			if err != nil {
				return toolError(err.Error())
			}
			type r struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Color       int    `json:"color"`
				Position    int    `json:"position"`
				Hoist       bool   `json:"hoist"`
				Mentionable bool   `json:"mentionable"`
				Permissions int64  `json:"permissions"`
				Managed     bool   `json:"managed"`
			}
			result := make([]r, 0, len(roles))
			for _, role := range roles {
				result = append(result, r{
					ID:          role.ID,
					Name:        role.Name,
					Color:       role.Color,
					Position:    role.Position,
					Hoist:       role.Hoist,
					Mentionable: role.Mentionable,
					Permissions: role.Permissions,
					Managed:     role.Managed,
				})
			}
			return resultJSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("create_role",
			mcp.WithDescription("Create a new role"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Role name")),
			mcp.WithNumber("color", mcp.Description("Role color as decimal integer (e.g. 3447003 for blue)")),
			mcp.WithBoolean("hoist", mcp.Description("Display role members separately in sidebar")),
			mcp.WithBoolean("mentionable", mcp.Description("Allow anyone to @mention this role")),
			mcp.WithString("permissions", mcp.Description("Permission bitfield string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error())
			}
			role, err := bot.GuildRoleCreate(guildID, &discordgo.RoleParams{
				Name:        name,
				Color:       intPtr(int(req.GetFloat("color", 0))),
				Hoist:       boolPtr(req.GetBool("hoist", false)),
				Mentionable: boolPtr(req.GetBool("mentionable", false)),
			})
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]interface{}{"id": role.ID, "name": role.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("edit_role",
			mcp.WithDescription("Edit an existing role"),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID to edit")),
			mcp.WithString("name", mcp.Description("New role name")),
			mcp.WithNumber("color", mcp.Description("New color as decimal integer")),
			mcp.WithBoolean("hoist", mcp.Description("Display separately")),
			mcp.WithBoolean("mentionable", mcp.Description("Allow mentions")),
			mcp.WithString("permissions", mcp.Description("Permission bitfield string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			roleID, err := req.RequireString("role_id")
			if err != nil {
				return toolError(err.Error())
			}
			rp := &discordgo.RoleParams{}
			if v := req.GetString("name", ""); v != "" {
				rp.Name = v
			}
			args := req.GetArguments()
			if _, ok := args["color"]; ok {
				rp.Color = intPtr(int(req.GetFloat("color", 0)))
			}
			if _, ok := args["hoist"]; ok {
				rp.Hoist = boolPtr(req.GetBool("hoist", false))
			}
			if _, ok := args["mentionable"]; ok {
				rp.Mentionable = boolPtr(req.GetBool("mentionable", false))
			}
			role, err := bot.GuildRoleEdit(guildID, roleID, rp)
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]interface{}{"id": role.ID, "name": role.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("delete_role",
			mcp.WithDescription("Delete a role (irreversible)"),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID to delete")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			roleID, err := req.RequireString("role_id")
			if err != nil {
				return toolError(err.Error())
			}
			err = bot.GuildRoleDelete(guildID, roleID)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("role deleted"), nil
		},
	)
}
