# Discord Voice Monitor

This application monitors one or more discord guilds and notifies a non e2e encrypted matrix protocol chat room when someone joins any voice channel. If "friends" are specified, the monitor will also output what if any games the listed friends are playing on the Valve Steam platform.

## Why?

Discord doesn't natively offer a way to notify you when someone joins a voice channel, so unless you are actively watching Steam / Discord or people are actively pinging you, it can be difficult to know when a spontaneously group gaming session has materialized while away from your desktop.

This provides an easy way of knowing what's up from any matrix protocol enabled device.

## Configuration

configuration is supplied with a `config.toml` file which accepts the following key-values:

* `discordtoken` the discord bot token to use. The application will report voice channel join events for every voice channel in every guild the bot has access to.
* `mxtoken` a matrix.org access token associated with a user account.
* `mxroom` the matrix.org room to post messages to. **The user specified by mxtoken must already be a member of the matrix room**. Format: `https://example.com/_matrix/client/unstable/rooms/!mQytUcdcuYTwtnpCmB%3Aexample.com`
* `friends` a list of Steam profile ids to check for active games. **This only works for users with game activity set to public**. No steam auth needed.
* `activityinterval` the minimum number of seconds between checking friends steam game activity. Higher numbers decrease message verbosity and decreases the number of requests to steam.

### Running as a Service

To run discord-voice-monitor as a service on systems running the systemd init system:

1. Clone and build the project to `/opt/discord-voice-monitor/`
2. copy `config.toml.dist` to `config.toml` and customize the settings as appropriate.
3. run `./add-service.sh` to add the service to systemd and start running the application.

### Logging

Basic logs will be printed to stdout and will be available via `journalctl -f -u discord-voice-monitor.service` when running via systemd.
