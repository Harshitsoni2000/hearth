# Hearth

**Hearth** is a zero-dependency LAN media server written in pure Go. Point
it at a directory of media and every device on the network — a laptop, a
phone, the living-room TV — gets a browsable library and a direct stream
into VLC, with no transcoding, no decoding, and no waiting for a whole
file to load. Each viewer seeks and pauses independently; there's no
shared playback state to synchronize.

Deployment is one binary: copy `hearth` onto a machine on the LAN, run it,
done.

## Build

```
make build
```

Produce `./hearth` binary in repo root.

## Run

```
./hearth -dir /path/to/media-directory
```

Flags:

- `-dir` — directory to serve media from. Default `.` (current directory).
- `-port` — port to listen on. Default `8080`.
- `-addr` — address to listen on. Default `0.0.0.0` (all interfaces).

## Endpoints

| Route | Returns |
|---|---|
| `GET /` | HTML tree view of the media library — folders collapsed by default, each file has a copy-link button |
| `GET /api/files` | JSON listing of the same tree, for scripting/automation |
| `GET /media/{path...}` | The raw file, streamed with `Range` support |

## Watching in VLC

1. On any device on the LAN, open a browser to `http://<server-ip>:8080/`.
2. Expand folders to find the file, tap the copy button next to it — this
   copies the file's full stream URL to the clipboard (works over plain HTTP,
   no HTTPS required).
3. In VLC: Media → Open Network Stream (Cmd+N on Mac), paste, hit Play.

Server prints the base URL for each LAN interface on startup, in case you'd
rather type a URL by hand: `http://<server-ip>:8080/media/<path>`.

### Faststart note

If `.mp4` file seek slow (buffer whole file just to jump), moov atom sit at end
of file. Fix once, locally:

```
ffmpeg -i in.mp4 -c copy -movflags +faststart out.mp4
```

`.mkv` file generally fine, no faststart needed.

## Not included yet

No sync / watch-together, no transcoding, no authentication — by design.
This is a LAN tool: it trusts the network it's running on.
