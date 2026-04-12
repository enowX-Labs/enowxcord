package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerWebhookTools(s *server.MCPServer, d *Discord) {
	bot := d.Session
	guildID := d.GuildID

	s.AddTool(
		mcp.NewTool("list_webhooks",
			mcp.WithDescription("List all webhooks in the server"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			webhooks, err := bot.GuildWebhooks(guildID)
			if err != nil {
				return toolError(err.Error())
			}
			type wh struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				ChannelID string `json:"channel_id"`
			}
			result := make([]wh, 0, len(webhooks))
			for _, w := range webhooks {
				result = append(result, wh{ID: w.ID, Name: w.Name, ChannelID: w.ChannelID})
			}
			return resultJSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("create_webhook",
			mcp.WithDescription("Create a webhook for a channel"),
			mcp.WithString("channel_id", mcp.Required(), mcp.Description("Channel ID for the webhook")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Webhook name")),
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
			wh, err := bot.WebhookCreate(channelID, name, "")
			if err != nil {
				return toolError(err.Error())
			}
			return resultJSON(map[string]string{
				"id":   wh.ID,
				"name": wh.Name,
				"url":  fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", wh.ID, wh.Token),
			})
		},
	)

	s.AddTool(
		mcp.NewTool("delete_webhook",
			mcp.WithDescription("Delete a webhook"),
			mcp.WithString("webhook_id", mcp.Required(), mcp.Description("Webhook ID to delete")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			webhookID, err := req.RequireString("webhook_id")
			if err != nil {
				return toolError(err.Error())
			}
			err = bot.WebhookDelete(webhookID)
			if err != nil {
				return toolError(err.Error())
			}
			return mcp.NewToolResultText("webhook deleted"), nil
		},
	)
}
