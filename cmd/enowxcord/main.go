package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/enowx/enowxcord/internal/discord"
	"github.com/enowx/enowxcord/internal/tools/channel"
	"github.com/enowx/enowxcord/internal/tools/guild"
	"github.com/enowx/enowxcord/internal/tools/invite"
	"github.com/enowx/enowxcord/internal/tools/member"
	"github.com/enowx/enowxcord/internal/tools/message"
	"github.com/enowx/enowxcord/internal/tools/role"
	"github.com/enowx/enowxcord/internal/tools/webhook"
)

func main() {
	pool := discord.NewSessionPool()
	defer pool.CloseAll()

	s := server.NewMCPServer(
		"enowxcord",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	channel.Register(s)
	role.Register(s)
	member.Register(s)
	guild.Register(s)
	webhook.Register(s)
	invite.Register(s)
	message.Register(s)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("enowxcord MCP server starting on :%s (SSE)", port)

	sseServer := server.NewSSEServer(s,
		server.WithSSEEndpoint("/sse"),
		server.WithMessageEndpoint("/message"),
		server.WithKeepAlive(true),
		server.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			token := r.Header.Get("X-Discord-Token")
			guildID := r.Header.Get("X-Guild-ID")
			if token == "" || guildID == "" {
				return ctx
			}
			client, err := pool.Get(token, guildID)
			if err != nil {
				log.Printf("failed to create discord session: %v", err)
				return ctx
			}
			return discord.WithClient(ctx, client)
		}),
	)

	if err := sseServer.Start(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}
