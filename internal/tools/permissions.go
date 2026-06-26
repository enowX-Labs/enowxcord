package tools

// PermissionBits maps Discord permission flag values to their names.
// See https://discord.com/developers/docs/topics/permissions
var PermissionBits = map[int64]string{
	1 << 0:  "CREATE_INSTANT_INVITE",
	1 << 1:  "KICK_MEMBERS",
	1 << 2:  "BAN_MEMBERS",
	1 << 3:  "ADMINISTRATOR",
	1 << 4:  "MANAGE_CHANNELS",
	1 << 5:  "MANAGE_GUILD",
	1 << 6:  "ADD_REACTIONS",
	1 << 7:  "VIEW_AUDIT_LOG",
	1 << 8:  "PRIORITY_SPEAKER",
	1 << 9:  "STREAM",
	1 << 10: "VIEW_CHANNEL",
	1 << 11: "SEND_MESSAGES",
	1 << 12: "SEND_TTS_MESSAGES",
	1 << 13: "MANAGE_MESSAGES",
	1 << 14: "EMBED_LINKS",
	1 << 15: "ATTACH_FILES",
	1 << 16: "READ_MESSAGE_HISTORY",
	1 << 17: "MENTION_EVERYONE",
	1 << 18: "USE_EXTERNAL_EMOJIS",
	1 << 19: "VIEW_GUILD_INSIGHTS",
	1 << 20: "CONNECT",
	1 << 21: "SPEAK",
	1 << 22: "MUTE_MEMBERS",
	1 << 23: "DEAFEN_MEMBERS",
	1 << 24: "MOVE_MEMBERS",
	1 << 25: "USE_VAD",
	1 << 26: "CHANGE_NICKNAME",
	1 << 27: "MANAGE_NICKNAMES",
	1 << 28: "MANAGE_ROLES",
	1 << 29: "MANAGE_WEBHOOKS",
	1 << 30: "MANAGE_GUILD_EXPRESSIONS",
	1 << 31: "USE_APPLICATION_COMMANDS",
	1 << 32: "REQUEST_TO_SPEAK",
	1 << 33: "MANAGE_EVENTS",
	1 << 34: "MANAGE_THREADS",
	1 << 35: "CREATE_PUBLIC_THREADS",
	1 << 36: "CREATE_PRIVATE_THREADS",
	1 << 37: "USE_EXTERNAL_STICKERS",
	1 << 38: "SEND_MESSAGES_IN_THREADS",
	1 << 39: "USE_EMBEDDED_ACTIVITIES",
	1 << 40: "MODERATE_MEMBERS",
	1 << 44: "USE_SOUNDBOARD",
	1 << 46: "USE_EXTERNAL_SOUNDS",
	1 << 47: "SEND_VOICE_MESSAGES",
}

func DescribePermissions(value int64) []string {
	var names []string
	for bit, name := range PermissionBits {
		if value&bit != 0 {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return []string{"NONE"}
	}
	return names
}
