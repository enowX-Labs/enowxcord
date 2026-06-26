package event

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
		mcp.NewTool("list_scheduled_events",
			mcp.WithDescription("List all scheduled events in the server"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			events, err := bot.GuildScheduledEvents(guildID, true)
			if err != nil {
				return tools.Error(err.Error())
			}
			result := make([]map[string]any, 0, len(events))
			for _, e := range events {
				result = append(result, summarize(e))
			}
			return tools.JSON(result)
		},
	)

	s.AddTool(
		mcp.NewTool("create_scheduled_event",
			mcp.WithDescription("Create a scheduled event. For voice/stage events set channel_id; for external events set location and end_time."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Event name (1-100 chars)")),
			mcp.WithString("start_time", mcp.Required(), mcp.Description("Start time in RFC3339 format (e.g. 2026-07-01T18:00:00Z)")),
			mcp.WithString("entity_type", mcp.Required(), mcp.Description("'voice', 'stage', or 'external'")),
			mcp.WithString("description", mcp.Description("Event description")),
			mcp.WithString("channel_id", mcp.Description("Voice/stage channel ID (required for voice/stage)")),
			mcp.WithString("location", mcp.Description("Physical location (required for external)")),
			mcp.WithString("end_time", mcp.Description("End time in RFC3339 (required for external)")),
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
			startStr, err := req.RequireString("start_time")
			if err != nil {
				return tools.Error(err.Error())
			}
			start, err := time.Parse(time.RFC3339, startStr)
			if err != nil {
				return tools.Errorf("invalid start_time, expected RFC3339: %v", err)
			}
			entityType, err := req.RequireString("entity_type")
			if err != nil {
				return tools.Error(err.Error())
			}
			params := &discordgo.GuildScheduledEventParams{
				Name:               name,
				Description:        req.GetString("description", ""),
				ScheduledStartTime: &start,
				PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
			}
			switch entityType {
			case "voice":
				params.EntityType = discordgo.GuildScheduledEventEntityTypeVoice
				params.ChannelID = req.GetString("channel_id", "")
				if params.ChannelID == "" {
					return tools.Error("channel_id is required for voice events")
				}
			case "stage":
				params.EntityType = discordgo.GuildScheduledEventEntityTypeStageInstance
				params.ChannelID = req.GetString("channel_id", "")
				if params.ChannelID == "" {
					return tools.Error("channel_id is required for stage events")
				}
			case "external":
				params.EntityType = discordgo.GuildScheduledEventEntityTypeExternal
				location := req.GetString("location", "")
				if location == "" {
					return tools.Error("location is required for external events")
				}
				params.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{Location: location}
				endStr := req.GetString("end_time", "")
				if endStr == "" {
					return tools.Error("end_time is required for external events")
				}
				end, perr := time.Parse(time.RFC3339, endStr)
				if perr != nil {
					return tools.Errorf("invalid end_time, expected RFC3339: %v", perr)
				}
				params.ScheduledEndTime = &end
			default:
				return tools.Error("entity_type must be 'voice', 'stage', or 'external'")
			}
			e, err := bot.GuildScheduledEventCreate(guildID, params)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(summarize(e))
		},
	)

	s.AddTool(
		mcp.NewTool("edit_scheduled_event",
			mcp.WithDescription("Edit a scheduled event (name, description, or status)"),
			mcp.WithString("event_id", mcp.Required(), mcp.Description("Scheduled event ID")),
			mcp.WithString("name", mcp.Description("New name")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("status", mcp.Description("New status: 'scheduled', 'active', 'completed', or 'canceled'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			eventID, err := req.RequireString("event_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			params := &discordgo.GuildScheduledEventParams{}
			if v := req.GetString("name", ""); v != "" {
				params.Name = v
			}
			if v := req.GetString("description", ""); v != "" {
				params.Description = v
			}
			switch req.GetString("status", "") {
			case "scheduled":
				params.Status = discordgo.GuildScheduledEventStatusScheduled
			case "active":
				params.Status = discordgo.GuildScheduledEventStatusActive
			case "completed":
				params.Status = discordgo.GuildScheduledEventStatusCompleted
			case "canceled":
				params.Status = discordgo.GuildScheduledEventStatusCanceled
			}
			e, err := bot.GuildScheduledEventEdit(guildID, eventID, params)
			if err != nil {
				return tools.Error(err.Error())
			}
			return tools.JSON(summarize(e))
		},
	)

	s.AddTool(
		mcp.NewTool("delete_scheduled_event",
			mcp.WithDescription("Delete a scheduled event"),
			mcp.WithString("event_id", mcp.Required(), mcp.Description("Scheduled event ID")),
			mcp.WithDestructiveHintAnnotation(true),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bot, guildID, errResult := tools.FromContext(ctx)
			if errResult != nil {
				return errResult, nil
			}
			eventID, err := req.RequireString("event_id")
			if err != nil {
				return tools.Error(err.Error())
			}
			if err = bot.GuildScheduledEventDelete(guildID, eventID); err != nil {
				return tools.Error(err.Error())
			}
			return mcp.NewToolResultText("scheduled event deleted"), nil
		},
	)
}

func summarize(e *discordgo.GuildScheduledEvent) map[string]any {
	m := map[string]any{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"status":      int(e.Status),
		"entity_type": int(e.EntityType),
		"start_time":  e.ScheduledStartTime.Format(time.RFC3339),
		"user_count":  e.UserCount,
	}
	if e.ChannelID != "" {
		m["channel_id"] = e.ChannelID
	}
	if e.EntityMetadata.Location != "" {
		m["location"] = e.EntityMetadata.Location
	}
	return m
}
