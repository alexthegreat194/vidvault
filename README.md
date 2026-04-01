# VIDVAULT

A small **local HTTP server** that turns a directory on your machine into a **video gallery**. It scans for supported video files, serves them with range requests (so scrubbing works in the browser), and includes a web UI for browsing, organizing into folders, uploading, and lightbox playback.

There is **no authentication**: anyone who can reach the bind address can list and stream files under the chosen root. Run it only on trusted networks or `localhost`.

## Requirements

- [Go](https://go.dev/dl/) **1.24.5** or newer (see `go.mod`)
- A modern desktop browser (the UI uses standard HTML/CSS/JS)

Optional, for formatting frontend JavaScript:

- [Node.js](https://nodejs.org/) — then `npm install` and `npm run format`

## Quick start

```bash
go build -o vidvault ./src
./vidvault
```

This serves the **current directory** on port **8765**, prints local and LAN URLs, and tries to open the app in your default browser.

Serve a specific folder and port:

```bash
./vidvault /path/to/media --port 9000
# shorthand
./vidvault ~/Videos -p 9000
```

The server listens on **all interfaces** (`0.0.0.0`), so other devices on your LAN can open the “network” URL shown at startup.

## Features

| Area | What you get |
|------|----------------|
| **Gallery** | Grid or list view, search filter, sort by name / folder / extension |
| **Folders** | Sidebar navigation; create folders; drag-and-drop move onto a folder; delete folder (moves contained files to root, with rename if names collide) |
| **Player** | Click a card to open a modal; prev/next and keyboard arrows; video served with `Accept-Ranges` for seeking |
| **Selection** | “Select” mode, select all / clear, move many files in one go |
| **Upload** | Modal upload with optional destination folder or new folder name; only allowed video types are accepted server-side |
| **Warnings** | Folders that contain non-video files are flagged in the UI (those files are not shown in the gallery) |

## Supported video formats

Files must use one of these extensions (case-insensitive):

`.webm`, `.mp4`, `.mpv`, `.mkv`, `.mov`, `.avi`, `.m4v`, `.ogv`

## HTTP API

All JSON bodies use `Content-Type: application/json` unless noted.

| Method & path | Description |
|---------------|-------------|
| `GET /` | Single-page HTML app (embedded assets). |
| `GET /api/videos` | JSON array of video objects (see below). |
| `GET /api/folders` | JSON array of `{ "name": string, "has_other_files": boolean }` for directories under the root. |
| `POST /api/mkdir` | Body: `{ "folder": "relative/path" }`. Creates the directory under the media root. |
| `POST /api/rmdir` | Body: `{ "folder": "relative/path" }`. Moves **files** inside that folder to the media root (renaming on conflict), then removes the folder tree. |
| `POST /api/move` | Body: `{ "path": "relative/file.mp4", "dest_folder": "target" }`. `dest_folder` may be empty or `/` for root. Creates destination dirs as needed. |
| `POST /api/upload` | `multipart/form-data` with fields `file` (one or more) and `folder` (destination path segment). Response: JSON array of `{ "name": "...", "error": "..." }` per file (`error` omitted on success). Max form memory ~512 MiB. |
| `GET /video?path=…` | Stream a file under the root. Query `path` is URL-encoded, slash-separated relative path. Returns 403 on path traversal; uses correct `Content-Type` and range support. |

**Video JSON shape** (`GET /api/videos`):

```json
{
  "name": "clip.mp4",
  "path": "subfolder/clip.mp4",
  "folder": "subfolder",
  "ext": ".mp4"
}
```

Root-level files use `"folder": "/"`.

## Project layout

| Path | Role |
|------|------|
| `go.mod` | Go module definition (`vidvault`). |
| `src/main.go` | CLI flags, resolve media root, print URLs, open browser, start HTTP server. |
| `src/server.go` | Routes, API handlers, video streaming. |
| `src/video.go` | Supported extensions, `Video` struct, directory scan. |
| `src/template.go` | `embed` of `src/web/*` and assembly of the single HTML document. |
| `src/browser.go` | LAN IP hint and OS-specific default browser launch. |
| `src/web/head.html`, `body.html`, `foot.html` | Page structure and markup. |
| `src/web/styles.css` | Styling. |
| `src/web/app.js` | Client logic (fetch APIs, UI). |
| `package.json` | `npm run format` runs Prettier on `src/web/**/*.js`. |

The shipped UI is **not** loaded from disk at runtime: it is concatenated from the embedded `src/web/` files at compile time. Change those files, then `go build ./src` again.

## Development

```bash
go build -o vidvault ./src
./vidvault ./some-test-media -p 8765
```

Format JavaScript:

```bash
npm install
npm run format
```
