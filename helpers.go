package main

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func resultJSON(data interface{}) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func toolError(msg string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(msg), nil
}

func intPtr(i int) *int    { return &i }
func boolPtr(b bool) *bool { return &b }

func describePermissionBits(value int64) []string {
	permMap := map[int64]string{
		1024:      "VIEW_CHANNEL",
		2048:      "SEND_MESSAGES",
		4096:      "SEND_TTS_MESSAGES",
		8192:      "MANAGE_MESSAGES",
		16384:     "EMBED_LINKS",
		32768:     "ATTACH_FILES",
		65536:     "READ_MESSAGE_HISTORY",
		16:        "MANAGE_CHANNELS",
		32:        "MANAGE_GUILD",
		2:         "KICK_MEMBERS",
		4:         "BAN_MEMBERS",
		8:         "ADMINISTRATOR",
		1048576:   "CONNECT",
		2097152:   "SPEAK",
		4194304:   "MUTE_MEMBERS",
		268435456: "MANAGE_ROLES",
		536870912: "MANAGE_WEBHOOKS",
		131072:    "MENTION_EVERYONE",
		262144:    "USE_EXTERNAL_EMOJIS",
	}
	var names []string
	for bit, name := range permMap {
		if value&bit != 0 {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return []string{"NONE"}
	}
	return names
}
