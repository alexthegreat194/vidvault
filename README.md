# Vidvault

Local **Go** HTTP server that turns a directory into a **video gallery**: it discovers supported video files, streams them with range requests (so seeking works in the browser), and serves a small web UI for browsing, organizing into folders, bulk selection/move, upload, and lightbox playback. The UI is embedded at compile time from `src/web/`.

**Security:** there is no authentication. Anyone who can reach the listen address can list and stream files under the media root. Use only on trusted networks or `localhost`.

## Screenshot

![Vidvault web UI (gallery)](docs/screenshot.png)

*Add a PNG (or change the path above) so this image appears—e.g. place a capture at `docs/screenshot.png` in the repository.*

## Requirements

- [Go](https://go.dev/dl/) **1.24.5**+ (see [`go.mod`](go.mod))
- A modern desktop browser (HTML, CSS, and client-side JavaScript)

Optional, for formatting frontend JavaScript:

- [Node.js](https://nodejs.org/) — run `npm install` and `npm run format`

## Run from source

```bash
go build -o vidvault ./src
./vidvault
```

- Serves the **current directory** on port **8765**
- Prints local and LAN URLs and opens the app in your default browser
- Listens on **all interfaces** (`0.0.0.0`) so other devices on your LAN can use the “network” URL

Serve a specific folder and port:

```bash
./vidvault /path/to/media --port 9000
# or
./vidvault ~/Videos -p 9000
```

## Run with Docker

The repo includes a multi-stage [`Dockerfile`](Dockerfile) and a helper script.

**Build and run (defaults: image `vidvault:local`, host port `8765`, host data directory `./data` → `/data` in the container):**

```bash
./scripts/docker-run.sh
```

Override with environment variables:

| Variable  | Default        | Role                                      |
| --------- | -------------- | ----------------------------------------- |
| `IMAGE`   | `vidvault:local` | Docker image name                       |
| `PORT`    | `8765`         | Host port mapped to the container’s `8765` |
| `DATA_DIR`| repo `./data`  | Host directory mounted as media root `/data` |

**Manual example:**

```bash
docker build -t vidvault:local .
docker run --rm -p 8765:8765 -v /path/to/your/videos:/data vidvault:local -p 8765 /data
```

The process inside the container uses port `8765`; map your host port with `-p HOST:8765` if you use something other than `docker-run.sh`.

## Features

| Area         | What you get                                                                 |
| ------------ | ----------------------------------------------------------------------------- |
| **Gallery**  | Grid or list view, search filter, sort by name / folder / extension            |
| **Folders**  | Sidebar; create folders; drag-and-drop onto a folder; delete folder (moves files to root, renames on collision) |
| **Player**   | Modal lightbox; prev/next and keyboard arrows; streaming with `Accept-Ranges` |
| **Selection**| Select mode, select all / clear, move many files at once                     |
| **Upload**   | Modal upload with optional destination or new folder; server validates video types only |
| **Warnings** | Folders that contain non-video files are flagged (those files are not listed)  |

## Supported video formats

Extensions (case-insensitive): `.webm` `.mp4` `.mpv` `.mkv` `.mov` `.avi` `.m4v` `.ogv`

## HTTP API

All JSON request bodies use `Content-Type: application/json` unless noted.

| Method & path      | Description |
| ------------------ | ----------- |
| `GET /`            | Single-page HTML (embedded assets). |
| `GET /api/videos`  | JSON array of video objects (shape below). |
| `GET /api/folders` | JSON array of `{ "name": string, "has_other_files": boolean }` — one entry per directory under the root; `name` is the slash-separated path relative to the media root. |
| `POST /api/mkdir`  | Body: `{ "folder": "relative/path" }` — create directory under the media root. |
| `POST /api/rmdir`  | Body: `{ "folder": "relative/path" }` — move files in that folder to the root (rename on conflict), remove the folder tree. |
| `POST /api/move`   | Body: `{ "path": "relative/file.mp4", "dest_folder": "target" }` — `dest_folder` may be `""` or `"/"` for root; creates parents as needed. |
| `POST /api/upload` | `multipart/form-data`: fields `file` (one or more) and `folder` (destination path). Response: JSON array of `{ "name", "error" }` per file (`error` omitted on success). `ParseMultipartForm` limit ≈ 512 MiB. |
| `GET /video?path=…` | Stream a file under the root; `path` is URL-encoded, slash-separated, relative. `403` on path traversal; correct `Content-Type` and range support. |

**`GET /api/videos` item shape:**

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
| ---- | ---- |
| [`go.mod`](go.mod) | Go module (`vidvault`). |
| [`Dockerfile`](Dockerfile) | Multi-stage image: build Go binary, Alpine runtime, default `CMD` serves `/data` on `8765`. |
| [`scripts/docker-run.sh`](scripts/docker-run.sh) | Build image and `docker run` with volume and port mapping. |
| [`package.json`](package.json) | `npm run format` — Prettier on `src/web/**/*.js`. |
| `src/main.go` | CLI flags, resolve media root, print URLs, open browser, start server. |
| `src/server.go` | Routes, API handlers, video streaming. |
| `src/video.go` | Supported extensions, `Video` struct, directory walk. |
| `src/template.go` | `embed` of `src/web/*` and single HTML document assembly. |
| `src/browser.go` | LAN IP and OS default browser. |
| `src/web/head.html`, `body.html`, `foot.html` | Page structure. |
| `src/web/styles.css` | Styling. |
| `src/web/app.js` | Client: API calls, UI. |

Rebuilding the binary (`go build -o vidvault ./src`) is required after changing `src/web/` or Go sources—the UI is not read from disk at runtime.

## Development

```bash
go build -o vidvault ./src
./vidvault ./some-test-media -p 8765
```

Format JavaScript with Prettier:

```bash
npm install
npm run format
```
