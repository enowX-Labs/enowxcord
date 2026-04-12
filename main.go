package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	discord, err := NewDiscord()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to discord: %v\n", err)
		os.Exit(1)
	}
	defer discord.Close()

	s := server.NewMCPServer(
		"enowxcord",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	registerChannelTools(s, discord)
	registerRoleTools(s, discord)
	registerMemberTools(s, discord)
	registerServerTools(s, discord)
	registerWebhookTools(s, discord)
	registerInviteTools(s, discord)
	registerMessageTools(s, discord)
	registerAdvancedChannelTools(s, discord)

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
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
