# SUPERLOUD Discord Bot

A Discord bot that yells back when you yell at it. Port of the original Ruby IRC bot.

## Prerequisites

- Go 1.22+
- A Discord account

## Discord Developer Portal Setup

1. Go to [discord.com/developers/applications](https://discord.com/developers/applications) and click **New Application**
2. Name your application -- **this name becomes the bot's display name** in Discord. You can change it later under the **Bot** tab. You can also set per-server nicknames in Discord itself (Server Settings > Members > right-click bot > Change Nickname).
3. Go to the **Bot** tab, click **Reset Token**, and copy the token. You'll need this later.
4. Still on the Bot tab, scroll down to **Privileged Gateway Intents** and enable **MESSAGE CONTENT**. This is required for the bot to read message text for prefix commands and loudness detection.
5. Go to **OAuth2 > URL Generator**:
   - **Scopes**: select `bot` and `applications.commands`
   - **Bot Permissions**: select `Send Messages`, `Read Message History`, `View Channels`
   - Copy the generated URL and open it in your browser to invite the bot to your server

## Bot Configuration

1. Copy the example config:
   ```
   cd discord
   cp config/config.yml.example config/config.yml
   ```

2. Edit `config/config.yml`:

   | Key | Description |
   |-----|-------------|
   | `channel_id` | Channel where the bot sends its startup greeting. Get this by enabling Developer Mode (User Settings > Advanced > Developer Mode), then right-click a channel > Copy Channel ID. |
   | `admin_ids` | List of Discord user IDs who can run admin commands (right-click user > Copy User ID). |
   | `db_path` | Path to the SQLite database (default: `db/louds.db`). |

   The other paths (`rps_path`, `ignores_path`, `whitelist_path`, `aliases_path`) have sensible defaults.

3. Set the bot token as an environment variable:
   ```
   export DISCORD_TOKEN=your_token_here
   ```

**Channel scope**: The bot responds in ALL channels it can see in the server. To restrict it to specific channels, use Discord's role permissions to limit which channels the bot role can access.

## Database Setup

**Fresh start**: The bot auto-creates `db/louds.db` with the correct schema on first run. No setup needed.

**Migrating from the IRC bot**: The Go bot uses the exact same SQLite schema as the Ruby bot. No migration needed -- just point the bot at the file:

- **Option A**: Copy to the default location:
  ```
  cp /path/to/louds.db discord/db/louds.db
  ```
- **Option B**: Set `db_path` in config.yml to the existing file:
  ```yaml
  db_path: "../louds.db"
  # or absolute:
  db_path: "/home/user/superloud/louds.db"
  ```

Old messages with IRC nick authors display as-is; new Discord messages store Discord display names.

**If the IRC bot is still running**: Copy the database rather than sharing it. SQLite doesn't handle concurrent writers from separate processes safely.

## Build & Run

```
cd discord
go build -o superloud ./cmd/superloud/
export DISCORD_TOKEN=your_token
./superloud
```

To use a custom config path:

```
./superloud /path/to/config.yml
```

## Commands

All prefix commands use `!` and must be UPPERCASE. Slash commands (`/command`) are also available.

| Prefix | Slash | Description |
|--------|-------|-------------|
| `!DONGME` | `/dongme` | Shows how much of a man you are |
| `!REDONGME` | `/redongme` | Reroll your dong with danger risk |
| `!SIZEME` | `/sizeme` | Tells you if you are worth anything to society |
| `!SIZE [user]` | `/size <user>` | Check someone else's size |
| `!BIGGESTDONG` | `/biggestdong` | Who has the biggest dong today |
| `!DONGWINNERS [n]` | `/dongwinners [count]` | Show the top dong winners (default 2) |
| `!DONGRANKME` | `/dongrankme` | Show your relative worth |
| `!DWALL` | `/dwall` | Show everybody's rank |
| `!UPVOTE` | `/upvote` | Vote the current message +1 |
| `!DOWNVOTE` | `/downvote` | Vote the current message -1 |
| `!SCORE` | `/score` | Show the last message's score |
| `!OMAKASE [tool]` | `/omakase [tool]` | Makes tools really great |
| `!RPS <object>` | `/rps <object>` | Play Rock Paper Scissors (extended edition) |
| `!HELP [cmd]` | `/help [command]` | Get help with commands |
| -- | `/admin refresh_ignores` | Reload the ignore list (admin only) |
| -- | `/admin refresh_aliases` | Reload the alias list (admin only) |

The bot also responds to @mentions with a random loud message.

## Optional Configuration Files

| File | Description |
|------|-------------|
| `config/ignores.txt` | One regex pattern per line for users to ignore |
| `config/whitelist.txt` | If present, ONLY matching users are allowed |
| `config/aliases.yml` | Map alt Discord IDs to a primary ID to prevent dong cheating |
| `rps/rps.yml` | RPS game objects and fight messages (pre-configured) |
