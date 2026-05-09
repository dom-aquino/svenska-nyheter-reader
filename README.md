# svenska-nyheter-reader

> This project is primarily a test of how [Claude Code](https://claude.ai/code) works. The actual functionality — a Swedish news reader — is the vehicle for that experiment.

A terminal-based CLI for reading Swedish news from [8 Sidor](https://8sidor.se), written in Go.

## Features

- Browse the latest articles from the 8 Sidor RSS feed
- Navigate the article list with arrow keys
- Read full articles in a centered, scrollable reader
- Navigate word-by-word (`w` / `b`) or line-by-line (`j` / `k`)
- Press `space` on any word to look up its English translation via the DeepL API

## Usage

```bash
go run .
```

To enable word translation, set your DeepL API key (free tier available at [deepl.com](https://www.deepl.com/pro#developer)):

```bash
DEEPL_API_KEY=your-key go run .
```

## Controls

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate article list |
| `enter` | Open article |
| `w` / `b` | Next / previous word |
| `j` / `k` | Next / previous line |
| `space` | Translate highlighted word |
| `esc` | Close popup / go back |
| `q` | Quit |
