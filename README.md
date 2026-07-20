# Hearth

Hearth be LAN media server, write in Go. It serve movie file over HTTP so VLC on
any device in house can stream them straight from disk — no transcode, no decode,
no whole-file load into memory, just byte range sent over wire. Each device play
independent: pause here, seek there, no sync between them.

## Build

```
make build
```

Produce `./hearth` binary in repo root.

## Run

```
./hearth -dir /path/to/movies
```

Flags:

- `-dir` — directory to serve media from. Default `.` (current directory).
- `-port` — port to listen on. Default `8080`.
- `-addr` — address to listen on. Default `0.0.0.0` (all interfaces).

## Watching in VLC

1. Open VLC.
2. Go Media → Open Network Stream (Cmd+N on Mac).
3. Paste URL: `http://<server-ip>:8080/media/<filename>`
4. Hit Play.

Server print base URL for each LAN interface on startup — copy one of those and
append `/media/<filename>` to it.

Open `/` in browser (plain `http://<server-ip>:8080/`) to get plaintext list of
every playable URL under `-dir`.

### On the TV

Install VLC app on TV. Open Network Stream same way, paste same URL, play.

### Faststart note

If `.mp4` file seek slow (buffer whole file just to jump), moov atom sit at end
of file. Fix once, locally:

```
ffmpeg -i in.mp4 -c copy -movflags +faststart out.mp4
```

`.mkv` file generally fine, no faststart needed.

## Not included yet

No web UI, no sync / watch-together, no transcoding — by design.
