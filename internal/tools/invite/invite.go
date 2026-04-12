package invite

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/tools"
)

func Register(s *server.MCPServer, bot *discordgo.Session, guildID string) {
	s.AddTool(
		mcp.NewTool("list_invites",
			mcp.WithDescription("List all active invites for the server"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			invites, err := bot.GuildInvites(guildID)
			if err != nil {
				return tools.Error(err.Error())
			}
			type entry struct {
				Code      string `json:"code"`
				ChannelID string `json:"channel_id"`
				Uses      int    `json:"uses"`
				MaxUses   int    `json:"max_uses"`
				MaxAge    int    `json:"max_age"`
				Temporary bool   `json:"temporary"`
				CreatedBy string `json:"created_by,omitempty"`
			}
			result := make([]entry, 0, len(invites))
			for _, i := range invites {
				e := entry{
					Code: i.Code, Uses: i.Uses, MaxUses: i.MaxUses,
					MaxAge: i.MaxAge, Temporary: i.Temporary,
				}
				if i.Channel != nil {
					e.ChannelID = i.Channel.ID
				}
				if i.Inviter != nil {
					e.CreatedBy = i.Inviter.Username
				}
				result = append(result, e)
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("create_invite",
			mcp.WithDescription("Create an invite link for a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID to create invite for")),
			mcp.WithNumber("max_age", mcp.Description("Invite expiry in seconds (0 = never, default 86400)")),
			mcp.WithNumber("max_uses", mcp.Description("Max uses (0 = unlimited)")),
			mcp.WithBoolean("temporary", mcp.Description("Grant temporary membership")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			channelID, err := req.RequireString("channel_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			inv, err := bot.ChannelInviteCreate(channelID, discordgo.Invite{
				MaxAge:    int(req.GetFloat("max_age", 86400)),
				MaxUses:   int(req.GetFloat("max_uses", 0)),
				Temporary: req.GetBool("temporary", false),
			})
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(map[string]string{"code": inv.Code, "url": "https://discord.gg/" + inv.Code})
		},
	)

	s.AddTool(
		mcp.NewTool("delete_invite",
			mcp.WithDescription("Delete/revoke an invite"),
			mcp.WithString("invite_code", mcp.Required(), mcp.Description("Invite code to delete")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			code, err := req.RequireString("invite_code")
			if err != nil {
				return tools.Error(err.Error())
			}
			if _, err = bot.InviteDelete(code); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("invite deleted"), nil
		},
	)
}
