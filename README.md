# Discord Voice Monitor

This application monitors one or more discord guilds and notifies a non e2e encrypted matrix.org channel when someone joins any voice channel.

configuration is supplied with a `config.toml` file which accepts the following key-values:

* `discordtoken` the discord bot token to use. The application will report voice channel join events for every voice channel in every guild the bot has access to.
* `mxtoken` a matrix.org access token associated with a user account.
* `mxroom` the matrix.org room to post messages to. Format: `https://example.com/_matrix/client/unstable/rooms/!mQytUcdcuYTwtnpCmB%3Aexample.com`

Notes:

It is assumed that the matrix account already is a member of the target room.
