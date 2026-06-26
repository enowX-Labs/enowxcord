# enowxcord

Multi-tenant MCP server for Discord server management. Exposes 73 tools for managing channels, roles, members, messages, reactions, threads, scheduled events, emojis, webhooks, and invites — usable from Claude Desktop, Cursor, OpenCode, or any MCP-compatible AI client.

Each user provides their own Discord bot token and guild ID via headers. No server-side credentials needed.

Tools are request-driven: the AI client calls them on demand. The server does not stream live Discord gateway events (no auto-replies or live moderation) — to read state, call the relevant `list_*`/`get_*` tool.

Built with Go, [mcp-go](https://github.com/mark3labs/mcp-go), and [discordgo](https://github.com/bwmarrin/discordgo).

## Quick Start

```bash
go build -o enowxcord ./cmd/enowxcord
./enowxcord
```

Server starts on port `8080` by default. Set `PORT` env var to change.

## Transports

Two transports are served on the same port:

| Endpoint | Transport | Use |
|---|---|---|
| `/mcp` | Streamable HTTP (current MCP standard) | Preferred for modern clients |
| `/sse` (+ `/message`) | HTTP+SSE (legacy) | For clients that only speak SSE |
| `/healthz` | Plain HTTP | Health check (returns `ok`) |

## Connect from AI Clients

Users provide their Discord bot token and guild ID via HTTP headers when connecting.

### Claude Desktop

Add to `~/.config/claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "enowxcord": {
      "url": "https://your-domain.com/mcp",
      "headers": {
        "X-Discord-Token": "your-bot-token",
        "X-Guild-ID": "your-guild-id"
      }
    }
  }
}
```

For clients that require SSE, use `https://your-domain.com/sse` instead.

### Cursor

MCP server URL: `https://your-domain.com/mcp` (or `/sse` for SSE-only clients).

Set headers `X-Discord-Token` and `X-Guild-ID` in the MCP configuration.

### OpenCode

```json
{
  "url": "https://your-domain.com/mcp",
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

73 tools across the categories below. Tools marked ⚠️ are destructive and annotated in the MCP schema so AI clients can warn before executing.

### Channels (9)

| Tool | Description |
|---|---|
| `get_channel` | Get details for a single channel |
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

### Roles (5)

| Tool | Description |
|---|---|
| `list_roles` | List all roles with permissions, colors, positions |
| `create_role` | Create role (name, color, hoist, mentionable) |
| `edit_role` | Edit role properties |
| `delete_role` | Delete role ⚠️ |
| `reorder_roles` | Reorder role positions (highest to lowest) |

### Members (12)

| Tool | Description |
|---|---|
| `list_members` | List server members (up to 100) |
| `get_member` | Get detailed member info |
| `search_members` | Search members by username/nickname prefix |
| `get_bot_user` | Get the bot's own user account |
| `edit_member` | Edit nickname, roles, mute, deaf |
| `add_role_to_member` | Add role to member |
| `remove_role_from_member` | Remove role from member |
| `kick_member` | Kick member ⚠️ |
| `ban_member` | Ban member ⚠️ |
| `unban_member` | Unban user |
| `timeout_member` | Timeout member (up to 28 days) |
| `list_bans` | List banned users |

### Server (9)

| Tool | Description |
|---|---|
| `get_server_info` | Server name, member count, boost level, features, vanity URL |
| `edit_server` | Edit server name and description |
| `set_server_icon` | Set the server icon (data URI or base64) |
| `list_emojis` | List custom emojis |
| `create_emoji` | Upload a custom emoji (data URI or base64) |
| `delete_emoji` | Delete a custom emoji ⚠️ |
| `list_integrations` | List server integrations |
| `get_audit_log` | Get recent audit log entries |
| `prune_members` | Prune inactive roleless members (supports dry_run) ⚠️ |

### Reactions (4)

| Tool | Description |
|---|---|
| `add_reaction` | Add a reaction to a message |
| `remove_reaction` | Remove the bot's (or a user's) reaction |
| `remove_all_reactions` | Remove all reactions, or all of one emoji ⚠️ |
| `list_reactions` | List users who reacted with an emoji |

### Threads (6)

| Tool | Description |
|---|---|
| `list_active_threads` | List active (non-archived) threads |
| `create_thread_from_message` | Create a thread attached to a message |
| `set_thread_state` | Archive/unarchive or lock/unlock a thread |
| `add_thread_member` | Add a member to a thread |
| `remove_thread_member` | Remove a member from a thread |
| `list_thread_members` | List members of a thread |

### Scheduled Events (4)

| Tool | Description |
|---|---|
| `list_scheduled_events` | List all scheduled events |
| `create_scheduled_event` | Create a voice/stage/external event |
| `edit_scheduled_event` | Edit name, description, or status |
| `delete_scheduled_event` | Delete a scheduled event ⚠️ |

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

### Messages (14)

| Tool | Description |
|---|---|
| `send_message` | Send text message |
| `send_embed` | Send rich embed (title, fields, author, image, thumbnail, etc.) |
| `send_dm` | Send a direct message to a user |
| `reply_message` | Send a message as a reply to another |
| `get_messages` | Read recent channel history |
| `get_message` | Get a single message by ID |
| `edit_message` | Edit a message the bot sent |
| `delete_message` | Delete a single message ⚠️ |
| `bulk_delete_messages` | Delete 2-100 messages ⚠️ |
| `crosspost_message` | Publish an announcement message |
| `pin_message` | Pin a message |
| `unpin_message` | Unpin a message |
| `list_pinned_messages` | List pinned messages in a channel |
| `create_thread` | Create a thread in a channel |

## Bot Permissions

The Discord bot needs these permissions (grant only what your use case requires):

- Manage Channels
- Manage Roles
- Kick Members
- Ban Members
- Moderate Members
- Manage Webhooks
- Create Invite
- Manage Messages
- Add Reactions
- Send Messages
- Read Message History
- Manage Emojis and Stickers
- Manage Events
- Manage Threads
- View Audit Log

**Privileged Intents**: Server Members Intent must be enabled in the Discord Developer Portal.

## License

MIT
