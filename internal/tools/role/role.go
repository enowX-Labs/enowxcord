package role

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

func intPtr(i int) *int    { return &i }
func boolPtr(b bool) *bool { return &b }

func Register(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("list_roles",
			mcp.WithDescription("List all roles in the server with their permissions, colors, and positions"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			roles, err := bot.GuildRoles(guildID)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Color       int    `json:"color"`
				Position    int    `json:"position"`
				Hoist       bool   `json:"hoist"`
				Mentionable bool   `json:"mentionable"`
				Permissions int64  `json:"permissions"`
				Managed     bool   `json:"managed"`
			}
			result := make([]entry, 0, len(roles))
			for _, r := range roles {
				result = append(result, entry{
					ID: r.ID, Name: r.Name, Color: r.Color, Position: r.Position,
					Hoist: r.Hoist, Mentionable: r.Mentionable,
					Permissions: r.Permissions, Managed: r.Managed,
				})
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("create_role",
			mcp.WithDescription("Create a new role"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Role name")),
			mcp.WithNumber("color", mcp.Description("Role color as decimal integer (e.g. 3447003 for blue)")),
			mcp.WithBoolean("hoist", mcp.Description("Display role members separately in sidebar")),
			mcp.WithBoolean("mentionable", mcp.Description("Allow anyone to @mention this role")),
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
			r, err := bot.GuildRoleCreate(guildID, &discordgo.RoleParams{
				Name:        name,
				Color:       intPtr(int(req.GetFloat("color", 0))),
				Hoist:       boolPtr(req.GetBool("hoist", false)),
				Mentionable: boolPtr(req.GetBool("mentionable", false)),
			})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]interface{}{"id": r.ID, "name": r.Name})
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
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			roleID, err := req.RequireString("role_id")
			if err != nil {
				return tools.Error(err.Error())
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
			r, err := bot.GuildRoleEdit(guildID, roleID, rp)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]interface{}{"id": r.ID, "name": r.Name})
		},
	)

	s.AddTool(
		mcp.NewTool("delete_role",
			mcp.WithDescription("Delete a role (irreversible)"),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID to delete")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			roleID, err := req.RequireString("role_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.GuildRoleDelete(guildID, roleID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("role deleted"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("reorder_roles",
			mcp.WithDescription("Reorder roles. Provide role IDs in the desired order (first = highest position). Roles not listed keep their relative order below."),
			mcp.WithArray("role_ids", mcp.Required(), mcp.Description("Role IDs from highest to lowest position"), mcp.Items(map[string]any{"type": "string"})),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			arr, ok := req.GetArguments()["role_ids"].([]any)
			if !ok || len(arr) == 0 {
				return tools.Error("role_ids is required and must be a non-empty array")
			}
			// Highest position number = top of the list. Assign descending positions.
			roles := make([]*discordgo.Role, 0, len(arr))
			pos := len(arr)
			for _, v := range arr {
				id, ok := v.(string)
				if !ok || id == "" {
					continue
				}
				roles = append(roles, &discordgo.Role{ID: id, Position: pos})
				pos--
			}
			updated, err := bot.GuildRoleReorder(guildID, roles)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Position int    `json:"position"`
			}
			result := make([]entry, 0, len(updated))
			for _, r := range updated {
				result = append(result, entry{ID: r.ID, Name: r.Name, Position: r.Position})
			}
			return tools.JSON(result)
		},
	)
}
