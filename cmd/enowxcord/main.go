package main

import (
	"fmt"
	"log"
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
	client, err := discord.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "discord: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	s := server.NewMCPServer(
		"enowxcord",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	bot := client.Session
	gid := client.GuildID

	channel.Register(s, bot, gid)
	role.Register(s, bot, gid)
	member.Register(s, bot, gid)
	guild.Register(s, bot, gid)
	webhook.Register(s, bot, gid)
	invite.Register(s, bot, gid)
	message.Register(s, bot)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("enowxcord MCP server starting on :%s (SSE)", port)

	sseServer := server.NewSSEServer(s,
		server.WithSSEEndpoint("/sse"),
		server.WithMessageEndpoint("/message"),
		server.WithKeepAlive(true),
	)

	if err := sseServer.Start(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}
