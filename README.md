# yt-transcript

Turns a YouTube link into one Markdown file an AI agent can read: title, chapters, description and a transcript where every block links to that exact moment in the video.

```sh
yt "https://www.youtube.com/watch?v=VIDEO_ID"
# prints the path of the file it wrote
```

## How it works

1. `yt-dlp` fetches the video metadata and the YouTube captions.
2. The captions are grouped into time blocks (30 s by default, set with `--window`).
3. One `.md` file is written to the library folder, with front matter (id, title, channel, date, duration), a chapter list, the description, and the transcript. Each timestamp is a `youtu.be/<id>?t=<seconds>s` link.
4. If a video has no captions, or you pass `--whisper`, the audio is transcribed with Groq Whisper instead.
5. A video already in the library is not fetched again unless you pass `--force`.

There is no built-in chat on purpose: the agent just needs a link to the context. Give your agent the file path and let it answer, citing the timestamp links.

## Install

Needs Go and [yt-dlp](https://github.com/yt-dlp/yt-dlp) on your PATH. Audio extraction for Whisper may also need ffmpeg.

```sh
go install github.com/cutemonstre/yt-transcript@latest
```

The Go tool names the binary `yt-transcript`. To get the short name, build it with `go build -o yt .` from a clone.

## Flags

`--library` (output folder), `--lang` (caption language, default `en`), `--window`, `--force`, `--whisper`.

## Environment variables

| name | purpose |
|---|---|
| `GROQ_API_KEY` | only for Whisper transcription |
| `YT_LIBRARY` | library folder (default `~/Documents/yt-library`) |

A `.env` file is read from the current directory, from `~/.config/yt-ask/.env`, and from next to the binary. Real environment variables win.

## Test

```sh
go build ./... && go vet ./... && go test ./...
```

Built by directing AI coding agents (Claude Code), with design, review and testing by the author.

## License

MIT
