# Voicybot TeamSpeak3 plugin

> Please note, this TeamSpeak3 plugin is highly experimental and is in active
> development! You will likely experience occasional crashes when building and
> installing this plugin and it is currently not aiming to be user-friendly.

This plugin is an attempted reimplementation of [TS3Bot](https://github.com/icedream/ts3bot)
as a TeamSpeak3 plugin in order to allow direct insertion of audio into TeamSpeak3
as opposed to having to use PulseAudio to feed it through a virtual input device
which introduces lag on unoptimized systems.

## Dependencies

To run this bot you will need:

- ffmpeg (needs to be executable through PATH)
- youtube-dl (needs to be executable through PATH, install using pip3 ideally to reduce playback waiting time)

Additionally, if you want to build the plugin yourself, you will need:

- Go 1.11+ (make sure Go Module support works!)

## Building the plugin

```
# Windows:
go build -v -buildmode=c-shared -o voicybot-ts3plugin.dll

# Linux:
go build -v -buildmode=c-shared -o voicybot-ts3plugin.so
```

Alternatively, running either [install.cmd](install.cmd) or [install.sh](install.sh)
will build and install the plugin into the current user's TS3Client plugin folder.

## Using the plugin

This plugin listens for text messages with commands to execute. Following commands
are supported:

- `play <url> [<password>]` - Makes the plugin play the given video or audio URL. You may have to wait 3-4 seconds for youtube-dl to do its job.
- `stop` - Stops current playback.
- `volume <value>` - Sets the volume at which to play audio. `<value>` may be any number between 0 and 100.

