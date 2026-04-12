# enowxcord

Multi-tenant MCP server for Discord server management. Exposes 40 tools for managing channels, roles, members, webhooks, invites, and messages — usable from Claude Desktop, Cursor, OpenCode, or any MCP-compatible AI client.

Each user provides their own Discord bot token and guild ID via headers. No server-side credentials needed.

Built with Go, [mcp-go](https://github.com/mark3labs/mcp-go), and [discordgo](https://github.com/bwmarrin/discordgo).

## Quick Start

```bash
go build -o enowxcord ./cmd/enowxcord
./enowxcord
```

Server starts on port `8080` by default. Set `PORT` env var to change.

## Connect from AI Clients

Users provide their Discord bot token and guild ID via HTTP headers when connecting.

### Claude Desktop

Add to `~/.config/claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "enowxcord": {
      "url": "https://your-domain.com/sse",
      "headers": {
        "X-Discord-Token": "your-bot-token",
        "X-Guild-ID": "your-guild-id"
      }
    }
  }
}
```

### Cursor

MCP server URL: `https://your-domain.com/sse`

Set headers `X-Discord-Token` and `X-Guild-ID` in the MCP configuration.

### OpenCode

```json
{
  "url": "https://your-domain.com/sse",
  "headers": {
    "X-Discord-Token": "your-bot-token",
    "X-Guild-ID": "your-guild-id"
  }
}
```

## Deploy with Docker

```bash
docker build -t enowxcord .
docker run -d -p 8080:8080 enowxcord
```

No environment variables needed on the server — each user authenticates via headers.

For Dokploy, Railway, etc.: deploy the container, set a domain, users connect via `https://your-domain.com/sse`.

## Headers

| Header | Required | Description |
|---|---|---|
| `X-Discord-Token` | Yes | User's Discord bot token |
| `X-Guild-ID` | Yes | Target Discord server ID |

Sessions are pooled per token+guild pair and reused across requests.

## Available Tools

### Channels (8)

| Tool | Description |
|---|---|
| `list_channels` | List all channels with types, categories, positions |
| `create_text_channel` | Create text channel (name, category, topic, nsfw, slowmode) |
| `create_voice_channel` | Create voice channel (name, category, bitrate, user limit) |
| `create_category` | Create channel category |
| `edit_channel` | Edit channel properties |
| `delete_channel` | Delete channel ⚠️ |
| `set_channel_permissions` | Set permission overrides for role/user on a channel |
| `sync_category_permissions` | Set permissions on category + sync to all children |

### Advanced Channels (4)

| Tool | Description |
|---|---|
| `create_announcement_channel` | Create news channel (crosspost support) |
| `create_stage_channel` | Create stage channel for audio events |
| `create_forum_channel` | Create forum channel |
| `create_forum_post` | Create post in a forum channel |

### Roles (4)

| Tool | Description |
|---|---|
| `list_roles` | List all roles with permissions, colors, positions |
| `create_role` | Create role (name, color, hoist, mentionable) |
| `edit_role` | Edit role properties |
| `delete_role` | Delete role ⚠️ |

### Members (10)

| Tool | Description |
|---|---|
| `list_members` | List server members (up to 100) |
| `get_member` | Get detailed member info |
| `edit_member` | Edit nickname, roles, mute, deaf |
| `add_role_to_member` | Add role to member |
| `remove_role_from_member` | Remove role from member |
| `kick_member` | Kick member ⚠️ |
| `ban_member` | Ban member ⚠️ |
| `unban_member` | Unban user |
| `timeout_member` | Timeout member (up to 28 days) |
| `list_bans` | List banned users |

### Server (3)

| Tool | Description |
|---|---|
| `get_server_info` | Server name, member count, boost level, features |
| `edit_server` | Edit server name and description |
| `list_emojis` | List custom emojis |

### Webhooks (3)

| Tool | Description |
|---|---|
| `list_webhooks` | List all webhooks |
| `create_webhook` | Create webhook (returns URL) |
| `delete_webhook` | Delete webhook ⚠️ |

### Invites (3)

| Tool | Description |
|---|---|
| `list_invites` | List active invites |
| `create_invite` | Create invite link |
| `delete_invite` | Revoke invite ⚠️ |

### Messages (5)

| Tool | Description |
|---|---|
| `send_message` | Send text message |
| `send_embed` | Send rich embed |
| `bulk_delete_messages` | Delete 2-100 messages ⚠️ |
| `pin_message` | Pin a message |
| `create_thread` | Create thread in channel |

⚠️ = destructive operation (annotated in MCP schema so AI clients can warn before executing)

## Bot Permissions

The Discord bot needs these permissions:

- Manage Channels
- Manage Roles
- Kick Members
- Ban Members
- Moderate Members
- Manage Webhooks
- Create Invite
- Manage Messages
- Send Messages
- Read Message History
- Manage Emojis and Stickers

**Privileged Intents**: Server Members Intent must be enabled in the Discord Developer Portal.

## License

MIT
