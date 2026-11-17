package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/cutemonstre/yt-transcript/internal/captions"
	"github.com/cutemonstre/yt-transcript/internal/library"
	"github.com/cutemonstre/yt-transcript/internal/markdown"
	"github.com/cutemonstre/yt-transcript/internal/transcribe"
	"github.com/cutemonstre/yt-transcript/internal/ytdlp"
)

func main() {
	libDir := flag.String("library", defaultLibrary(), "Directory the transcript .md files live in")
	lang := flag.String("lang", "en", "Caption language to fetch (yt-dlp --sub-langs value)")
	window := flag.Duration("window", 30*time.Second, "Transcript block size — one timestamp link per block")
	force := flag.Bool("force", false, "Re-fetch even if the video is already in the library")
	whisper := flag.Bool("whisper", false, "Skip YouTube captions and transcribe with Groq Whisper")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: yt [flags] <youtube-url>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *window <= 0 {
		log.Fatal("--window must be greater than zero")
	}
	url := flag.Arg(0)

	// Process environment wins. A cwd .env can override the global fallback,
	// while the global file makes the key available when yt runs anywhere.
	_ = godotenv.Load()
	if home, err := os.UserHomeDir(); err == nil {
		_ = godotenv.Load(filepath.Join(home, ".config", "yt-ask", ".env"))
	}
	if exe, err := os.Executable(); err == nil {
		_ = godotenv.Load(filepath.Join(filepath.Dir(exe), ".env"))
	}

	videoID, err := library.VideoID(url)
	if err != nil {
		log.Fatalf("Not a YouTube URL I understand: %v", err)
	}
	if !*force {
		if existing, ok := library.Find(*libDir, videoID); ok {
			fmt.Fprintln(os.Stderr, "Already in library (use --force to re-fetch)")
			fmt.Println(existing)
			return
		}
	}

	ctx := context.Background()
	fmt.Fprintln(os.Stderr, "Step 1/3: Fetching metadata…")
	meta, err := ytdlp.Metadata(ctx, url)
	if err != nil {
		log.Fatalf("Metadata fetch failed: %v", err)
	}
	fmt.Fprintf(os.Stderr, "  %s — %s (%s)\n", meta.Channel, meta.Title, meta.DurationString())

	var segs []transcribe.Segment
	source := ""
	if !*whisper {
		fmt.Fprintf(os.Stderr, "Step 2/3: Fetching YouTube captions (%s)…\n", *lang)
		raw, capSource, captionErr := ytdlp.Captions(ctx, url, *lang)
		switch {
		case captionErr == nil:
			segs, err = captions.Parse(raw)
			if err != nil {
				log.Fatalf("Caption parse failed: %v", err)
			}
			source = capSource
		case errors.Is(captionErr, ytdlp.ErrNoCaptions):
			fmt.Fprintln(os.Stderr, "  No captions on this video — falling back to Whisper")
		default:
			log.Fatalf("Caption fetch failed: %v", captionErr)
		}
	}

	if segs == nil {
		key := os.Getenv("GROQ_API_KEY")
		if key == "" {
			log.Fatal("This video needs Whisper transcription: set GROQ_API_KEY in .env or the environment")
		}
		fmt.Fprintln(os.Stderr, "Step 2/3: Downloading audio…")
		audioPath, audioErr := ytdlp.Audio(ctx, url)
		if audioErr != nil {
			log.Fatalf("Audio download failed: %v", audioErr)
		}
		defer os.RemoveAll(filepath.Dir(audioPath))

		const groqUploadLimit = 25 << 20
		if fi, statErr := os.Stat(audioPath); statErr == nil && fi.Size() > groqUploadLimit {
			log.Fatalf("Audio is %.1f MB, over Groq's 25 MB cap (~2h18m of speech). v1 has no chunking — pick a shorter video or wait for captions.",
				float64(fi.Size())/(1<<20))
		}

		tctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		segs, err = transcribe.Transcribe(tctx, key, audioPath)
		cancel()
		if err != nil {
			log.Fatalf("Transcription failed: %v", err)
		}
		source = "whisper-groq"
	}

	if len(segs) == 0 {
		log.Fatal("No speech found — nothing to write")
	}

	fmt.Fprintf(os.Stderr, "Step 3/3: Writing markdown (%d segments, source=%s)…\n", len(segs), source)
	doc := markdown.Render(meta, segs, source, *window)
	outPath, err := library.Write(*libDir, meta, doc)
	if err != nil {
		log.Fatalf("Library write failed: %v", err)
	}
	fmt.Println(outPath)
}

func defaultLibrary() string {
	if v := os.Getenv("YT_LIBRARY"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "yt-library")
}
