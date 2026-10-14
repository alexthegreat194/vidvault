# How Vidvault works

This is a map of the running system: the Go server, the on-disk media library, the JSON stores that live outside the media folder, and the Svelte UI. The source of truth is the code. Parts of the root `README.md` still describe an older single-file JavaScript UI (`src/web/app.js`); that UI is gone. The app is a Svelte 5 frontend built into `src/web/dist/` and embedded in the binary.

Vidvault is a local video gallery. You point one process at one directory (the **media root**). The server walks that tree, hashes every supported video, and serves a browser UI that browses, filters, plays, uploads, moves, tags, and groups those files. Nothing is transcoded. Playback is the original file, streamed with HTTP range requests so the browser can seek.

There is no user database and no multi-tenant model. Identity of a video, for favorites, tags, collections, and “have I seen this,” is the **SHA-256 of the file bytes**, not the path. Moving or renaming a file keeps those associations. Two identical files share one hash, which is also how duplicates are detected.

---

## 1. The two halves

```
Browser (Svelte 5, Vite)
        │  fetch /api/*   and   <video src="/video?path=…">
        ▼
Go HTTP server  (package main in src/)
        │
        ├── media root on disk     videos and folders you actually own
        └── OS config dir          favorites, tags, collections, seen files, batches
              ~/Library/Application Support/vidvault/   (macOS)
```

| Piece | Where | What it owns |
| --- | --- | --- |
| CLI and process | `src/main.go` | Flags, bind address, browser open, start the server |
| HTTP | `src/server.go` | Routes, scan orchestration, upload/move/rename/delete, streaming |
| Scan and files | `src/video/video.go`, `src/video/library.go` | Extensions, hashing, scan cache, in-memory library |
| Folders | `src/folders.go` | List, mkdir, rmdir, move |
| Favorites | `src/favorites.go` | `hash → true` |
| Tags | `src/tags.go` | Tag definitions and `hash → [tag id]` |
| Watch collections | `src/collections.go` | Named playlists of hashes |
| Seen / new | `src/seen_files.go` | First-seen time, reviewed time, batch id, keyed by hash |
| Batches | `src/upload_batches.go` | Upload, discovered, and migrated groups |
| PIN | `src/auth.go` | Optional cookie gate |
| UI assets | `src/template.go` | `//go:embed web/dist/*` |
| Frontend | `frontend/src/` | Gallery, sidebar, modals, stores |

The Go module is `vidvault`. The server type and most stores live in `package main`. Only the scan/library code is a separate package, `vidvault/src/video`, so the server can call it without an import cycle.

---

## 2. Startup

`main` parses flags, then the optional media directory. Go’s `flag` package stops at the first non-flag argument, so flags must come first:

```text
./vidvault -p 9000 -pin 1234 ~/Videos
```

| Flag | Default | Effect |
| --- | --- | --- |
| positional path | `.` | Media root. Resolved to an absolute directory; the process exits if it is not a directory. |
| `-p` | `8765` | TCP port. The process listens on `0.0.0.0`, so other machines on the LAN can connect. |
| `-d` | off | Debug logs (slog). Without it, info-level request lines still print. |
| `-disable-browser` | off | Skip opening the default browser. |
| `-pin` | empty | When set, almost every route requires an unlock cookie. |

`newServer` then:

1. Opens the five JSON stores under the OS user config directory (`os.UserConfigDir()/vidvault`).
2. Wires the seen-files store and the batch store to each other.
3. Runs a one-time migration that turns old unbatched seen-file records into hour-window batches (`MigrateSeenFiles`).
4. Creates an empty in-memory scan cache and video library.
5. Registers routes.
6. Starts a background scan (`startupScan`) and returns. HTTP begins serving immediately; the library is not ready until that scan finishes.

The startup scan walks the media root, hashes files that are not already cached by size and mtime, marks seen/new state, and replaces the in-memory library. When that finishes, `library.SetReady(true)`. Later `GET /api/videos` and `GET /api/videos/stream` calls reuse that library unless the client asks for `?rescan=1`.

The console prints a localhost URL, a LAN URL (`lanIP()` in `src/browser.go` picks a non-loopback IPv4), and the absolute media root.

### Docker

The image’s working directory and default command serve `/data`. Bind-mount the real library there. If the process’s media root is `/`, the walk includes the container filesystem (`bin`, `etc`, …). `docker-entrypoint.sh` turns `VIDVAULT_DEBUG`, `VIDVAULT_DISABLE_BROWSER`, and `VIDVAULT_PIN` into the same flags. See `notes/NOTES.md`.

---

## 3. The media root

One directory is the whole library. The server never writes outside it. Every path the API accepts is relative to that root, slash-separated in JSON (`sub/clip.mp4`), even on Windows.

Path checks follow the same pattern everywhere (stream, upload, move, rename, delete, mkdir, rmdir):

1. Clean the relative path and reject `..`.
2. Join it onto the absolute root.
3. `filepath.Abs` both sides and require the result to stay inside the root (`root + separator`, or exactly the root for a destination directory).

Supported extensions, case-insensitive, are `.webm`, `.mp4`, `.mpv`, `.mkv`, `.mov`, `.avi`, `.m4v`, `.ogv`. Anything else is invisible in the gallery. A folder that contains non-video files is still listed, with `has_other_files: true`, and the empty-gallery message warns about them.

Zero-byte files are skipped. A single unreadable file is logged and skipped; it does not abort the scan.

Root-level videos use `"folder": "/"`. A file at `trips/2024/clip.mp4` has `"folder": "trips/2024"`.

---

## 4. Discovering videos

### Scan

`ScanVideosWithCallback` (`src/video/video.go`) is `filepath.WalkDir` over the media root.

For each regular file with a known extension and a non-zero size:

1. Build a cache key from `size` and `modtime` (nanoseconds): `"12345:1710000000000000000"`.
2. If the in-memory `VideoScanCache` has that relative path with the same key, reuse the cached `Video` (no re-hash).
3. Otherwise SHA-256 the whole file (`hashFile`) and store the result in the cache.
4. Count it as found, and as either hashed or cached. Progress callbacks fire about every 25 files or 500 ms, whichever comes first, so a huge library does not flood the stream.

After the walk, results are sorted by relative path. Cache entries whose paths disappeared are pruned.

The cache lives only in process memory. Restarting the server hashes again, unless size and mtime match entries rebuilt during that same process. It is a scan accelerator, not a persistent index.

### Video object

```json
{
  "name": "clip.mp4",
  "path": "trips/clip.mp4",
  "folder": "trips",
  "ext": ".mp4",
  "size": 12345678,
  "modified": "2026-04-28T06:30:00Z",
  "hash": "2d7d4f4b…",
  "is_favorite": false,
  "is_new": true,
  "batch_id": "batch_ab12…",
  "tags": ["tag_e319…"]
}
```

`name`, `path`, `folder`, `ext`, `size`, `modified`, and `hash` come from the filesystem scan. `is_favorite`, `tags`, `is_new`, and `batch_id` are applied afterward (`enrichVideos` / `applyVideoMetadata`) from the JSON stores. They are not stored next to the file.

### In-memory library

`VideoLibrary` is a mutex-guarded `map[path]Video` plus a `ready` flag. `AllSorted()` returns a path-sorted snapshot.

- Startup scan, a full rescan, and the streaming rescan call `ReplaceAll`.
- Upload, move, and rename `Upsert` one entry (and remove the old path).
- Delete `Remove`s the path.

`GET /api/videos` without `rescan=1`, once ready, slices this map. It does not walk the disk. Optional `offset` and `limit` apply after the path sort. Metadata (favorite, tags, batch) is applied only to the slice being returned.

`GET /api/videos/stream` is newline-delimited JSON (`application/x-ndjson`). If the library is already ready and `rescan` is not set, it emits one `video` event per cached entry, then `scan_complete`. If it must scan, events are:

| `type` | `payload` |
| --- | --- |
| `scan_start` | `{ started_at }` |
| `scan_progress` | `{ found, hashed, cached }` |
| `video` | a full video object |
| `scan_complete` | `{ count, duration_ms }` |
| `scan_error` | `{ message }` |

Videos are emitted only after the scan and seen-file enrichment finish, not file-by-file during the walk. The client still ingests them as a stream and flushes into the gallery in 200 ms batches so the UI can paint before the whole body is parsed.

`GET /api/videos/random?count=N&exclude=hash,hash` shuffles the ready library, drops excluded hashes, and returns up to `count` (default 10, max 100). Discover mode uses this. If the library is not ready it returns `[]`.

---

## 5. Identity: hash versus path

| Concern | Key |
| --- | --- |
| Which file on disk, streaming, move, rename, delete | relative path |
| Favorite, tags, collection membership, seen/new, batch | SHA-256 of contents |
| Duplicate groups in the UI | same hash, more than one path |
| Scan cache | relative path + size + mtime |

Seen-file keys are `hash:<sha256>` when a hash exists, otherwise `path:<relative path>`. Normal videos always have a hash.

Because tags and favorites follow the hash, renaming `a.mp4` to `b.mp4` does not drop them. Replacing the bytes (a new encode of the “same” clip) produces a new hash, so it looks like a new file. Two byte-identical copies are one logical video for tags and two files for the gallery; the duplicate manager exists to delete the extras.

---

## 6. Seen files, “new,” and batches

These two stores answer “what arrived, and have I sorted it?”

### Seen files — `seen_files.json`

```json
{
  "baselined_roots": { "/abs/path/to/media": true },
  "entries": {
    "hash:abc…": {
      "first_seen_at": "2026-10-14T07:00:00Z",
      "reviewed_at": "",
      "batch_id": "batch_…"
    }
  }
}
```

`is_new` is true when the record exists and `reviewed_at` is empty.

**First scan of a media root** baselines it. Every file already on disk is recorded with `reviewed_at` set to now, so a library you already have does not light up as thousands of incoming videos. `baselined_roots` remembers that absolute root. A legacy `initialized: true` flag (from before per-root baselines) is treated as “already baselined” for any root.

**Later scans** only mark files whose hash is missing from `entries`. Those are new. They are attached to a discovered batch (below) and left unreviewed.

Reviewing (`POST /api/videos/new/clear`) sets `reviewed_at`. The body is `{ "all": bool, "paths": [], "hashes": [] }`. The gallery’s “Mark as resolved” sends the paths currently in the filtered batch view.

Forgetting (`POST /api/videos/new/forget`) deletes the seen-file records so the next scan treats those hashes as new again.

### Batches — `upload_batches.json`

A batch is a named group of arrivals, not a folder.

```json
{
  "id": "batch_<16 hex chars>",
  "name": "Oct 14, 12:35 AM",
  "created_at": "2026-10-14T07:35:00Z",
  "needs_sorting": true,
  "source": "upload"
}
```

Default name is the local timestamp, `Jan 2, 3:04 PM`. `source` is one of:

| Source | When |
| --- | --- |
| `upload` | The upload modal creates a batch, then posts each file with that `batch_id`. |
| `discovered` | Files showed up on disk between scans (copied in outside the app, or generated after the first baseline). New files within 30 minutes reuse the latest discovered batch; older than that, a new one is created. |
| `migrated` | One-time upgrade. Seen-file rows that had no `batch_id` are grouped by the hour of `first_seen_at`. A group that still has unreviewed files gets `needs_sorting: true`. |

`GET /api/batches` joins batch definitions with counts from seen-file records: `video_count` (rows with that batch id) and `new_count` (those with an empty `reviewed_at`). The list is newest first.

`PATCH /api/batches/{id}` can rename a batch or set `needs_sorting`. An empty name falls back to the timestamp name.

The sidebar lists batches. Choosing one sets `filters.activeBatchId`, and the gallery keeps only videos whose `batch_id` matches. The incoming banner appears when some batch has `needs_sorting && new_count > 0`. “View” applies that batch filter. “Sort” opens the incoming modal for the same batch so you can send the new files at a folder and tags together.

Counts come from every seen-file row with that batch id, including hashes no longer on disk. The gallery itself only shows files the latest scan actually found.

---

## 7. Other persisted state

All of these live in the user config directory, next to the media root only by coincidence. They survive moving the binary. They are **not** inside the media folder, so a Docker bind-mount of `/data` does not include them unless you also persist the container user’s config dir. Deleting `seen_files.json` makes the next scan baseline again (everything looks already reviewed) if that root was never baselined, or treat missing hashes as new if the root key is still present. To re-test “incoming,” delete the seen-file store or add files after the first successful scan. The README’s note about the macOS path is right: `~/Library/Application Support/vidvault/seen_files.json`.

| File | Contents | Key |
| --- | --- | --- |
| `favorites.json` | map of hash to `true` | content hash |
| `tags.json` | `{ tags: [{id,name,color}], assignments: { hash: [tagId] } }` | content hash |
| `watch_collections.json` | `{ collections: [{id,name,video_hashes,created_at,updated_at}] }` | content hash, order preserved |
| `seen_files.json` | baselines + per-hash first seen / reviewed / batch | content hash |
| `upload_batches.json` | batch list + `batches_migrated` | batch id |

Writes are mutex-guarded. Batch and seen-file stores write `path.tmp` and `rename` over the real file. Tag ids look like `tag_<hex>`. Collection ids look like `col_<hex>`. Tag colors are generated, not chosen in the API.

Favorites, tags, and collections do not care which folder a file sits in. A collection built from a folder (`POST /api/collections/from-folder`) copies the hashes of videos in that folder at that moment; later moves do not add or remove membership.

---

## 8. Folders

`GET /api/folders` walks the tree and returns every subdirectory, not only ones that contain videos:

```json
{ "name": "trips/2024", "has_other_files": false }
```

`POST /api/mkdir` with `{ "folder" }` creates the directory (`MkdirAll`), including parents.

`POST /api/rmdir` refuses the root. It moves files that lived in the deleted folder up to the media root, renaming on name collision, then removes the directory tree.

`POST /api/move` with `{ "path", "dest_folder" }` moves one file. `dest_folder` of `""` or `"/"` means the media root. Parents are created as needed. The handler updates the in-memory library: remove the old path, re-read the file at the new path (hash can come from cache if size and mtime match), re-apply seen/batch metadata, and upsert. The response is `{ path, old_path, video }` so the UI can patch one card instead of rescanning.

The sidebar folder list is the union of folders from this API and folders implied by video paths, plus a synthetic `__all__` row. Filtering compares `video.folder` to the active folder (`"/"` for the root).

---

## 9. Tags, favorites, collections

### Tags

- `GET /api/tags` returns definitions and the full assignment map.
- `POST /api/tags/create` `{ "name" }` adds a tag with a random color.
- `POST /api/tags/assign` `{ "hash", "tag_id", "assigned" }` toggles one hash.
- `POST /api/tags/assign/bulk` applies one tag to many hashes.
- `POST /api/tags/delete` removes the tag and strips it from every assignment.
- `POST /api/tags/videos` returns assignments for a set of hashes (used when the UI wants a narrower refresh).

The gallery filter is an AND across selected tag ids: a video must carry every selected tag. Counts in the sidebar are computed in the client from the loaded video list, not from a separate endpoint.

### Favorites

`POST /api/favorites/set` `{ "hash", "favorite" }` sets or clears one hash. The header’s favorites toggle filters to `is_favorite`. Discover mode can favorite the clip on screen with the same endpoint; the client also patches the in-memory video so the heart stays in sync without a rescan.

### Watch collections

A collection is an ordered list of hashes, a name, and timestamps. It is a playlist, not a folder.

- Create, rename, delete.
- `POST /api/collections/videos/set` adds or removes one hash.
- `POST /api/collections/videos/bulk` appends many hashes, skipping ones already present.
- `POST /api/collections/from-folder` creates a collection from every video currently in a folder.

The client resolves a collection by walking `video_hashes` and looking each hash up in the loaded library. The first loaded file with that hash wins. Hashes with no file left on disk are reported as missing. Opening a collection (`WatchCollectionModal`) plays that ordered set in a grid (1–4 columns). Column count, master volume, and mute are stored in `localStorage`, not on the server.

Sidebar collection counts only include hashes that still match a loaded video.

---

## 10. Upload, rename, delete

### Upload

`POST /api/upload` is `multipart/form-data`:

- `file` — one or more parts
- `folder` — destination relative to the root (`/` or empty means the root)
- `batch_id` — optional; must already exist

`ParseMultipartForm` allows about 512 MiB. The destination is created if needed, after the same traversal check as everywhere else. Each file is accepted only if its extension is in the video set (`uploadOne`). The original basename is kept. The response is an array of `{ name, error?, video? }`. On success the new file is hashed, enriched (which marks it new, because the root is already baselined), assigned to `batch_id` when one was sent, and upserted into the library.

The upload modal creates the batch first (`POST /api/batches` with source `upload`), then uploads into it. The UI therefore gets a sidebar batch immediately, with `needs_sorting: true`.

### Rename

`POST /api/rename` `{ "path", "name" }` renames within the same directory. The new name must be a single path segment. If you omit an extension, the old one is kept. If you include one, it must be a supported video extension. Destination must not already exist (409). The library entry moves from the old path to the new one. The hash does not change, so tags and favorites stay.

### Delete

`POST /api/delete` `{ "path" }` removes the file from disk and from the library. It does not scrub the hash from favorites, tags, collections, or seen files. A later identical re-upload is still the same hash. A duplicate group disappears from the UI once only one path with that hash remains, because groups are computed client-side from the loaded list.

---

## 11. Playback

`GET /video?path=<urlencoded relative path>` opens the file and calls `http.ServeContent`. That implementation:

- sets `Content-Type` from the extension (`application/octet-stream` if unknown),
- advertises `Accept-Ranges: bytes`,
- honors `Range` so `<video>` can seek without downloading the whole file.

There is no thumbnail pipeline and no transcode. Gallery cards are `<video>` elements pointed at the same URL, usually seeked a couple of seconds in for a preview (`HOVER_PREVIEW_SEEK_SEC` is 2). If the browser cannot decode the codec (common for some `.mkv` or `.avi` files), the tile stays dark and the lightbox may not play. Hover-preview audio and volume are local settings.

The lightbox (`VideoPlayerModal`) plays one file and steps prev/next through the **filtered** list, not the whole library. Arrow keys follow that order.

Discover mode is a separate full-screen feed. It asks `/api/videos/random` for batches of 8, remembers hashes it has already shown, and prefetches when you are within 3 clips of the end. If the random endpoint fails, it shuffles the already loaded library locally. IntersectionObserver plays the clip in view and pauses the others. Mute is local to that view.

---

## 12. PIN gate

With no `-pin`, every route is open. Anyone who can reach the listen address can list and stream the media root. That is the intended deployment: localhost or a trusted LAN.

With `-pin`:

- `GET /`, `GET /favicon.svg`, `GET /assets/…`, `GET /api/auth/status`, and `POST /api/auth/pin` stay public so the shell and the lock screen can load.
- Everything else, including `/video` and `/api/videos`, returns 401 until the cookie is valid.

Unlock compares SHA-256 of the submitted PIN to SHA-256 of the configured PIN in constant time. A match sets an `HttpOnly` cookie `vv_pin` whose value is HMAC-SHA256 of `"unlock:"` plus a generation counter, keyed by a random secret created at process start. The cookie lasts about ten years. `POST /api/auth/lock` increments the generation, which invalidates every previously issued cookie, and clears the cookie.

The frontend calls `/api/auth/status` before loading the library (`bootstrap` in `library.ts`). If a PIN is required and the cookie is missing, `PinGate` covers the app. A later 401 on any fetch (session locked from another action) clears the in-memory library and shows the gate again. The lock control is only rendered when a PIN is configured and the session is unlocked.

The PIN is a process flag, not a user account. Restarting with a different `-pin` does not migrate anything; the cookie secret is new, so old cookies fail closed.

---

## 13. Frontend

### Build

Vite compiles `frontend/` to `src/web/dist/`. `src/template.go` embeds that directory. `GET /` serves `index.html`. `GET /assets/…` and `GET /favicon.svg` serve the embedded files. Changing Svelte or CSS requires a frontend build before `go build`, or the binary keeps the previous UI.

Dev uses two processes. Vite on port 5173 proxies `/api`, `/video`, and `/favicon.svg` to `VIDVAULT_BACKEND` (default `http://localhost:8765`). The Go server should be started with `-disable-browser`. `npm run dev:remote` is `scripts/dev-remote.sh`.

`App.svelte` is the shell: pin gate, header, duplicate banner, incoming banner, sidebar, gallery, and one instance of each modal. Modals are always mounted and shown by stores, not by routes. There is no client-side router.

### Stores

| Store module | Holds |
| --- | --- |
| `lib/stores/library.ts` | Videos, folders, tags, collections, batches, filters, derived lists, load/refresh |
| `lib/stores/ui.ts` | Which modal is open, sidebar, discover mode, context menu, preview and watch-grid preferences |
| `lib/stores/selection.ts` | Select mode and the set of selected paths |
| `lib/stores/player.ts` | Path of the file open in the lightbox |
| `lib/api.ts` | `fetch` wrappers |
| `lib/auth.ts` | Pin-gate flags and the 401 interceptor |

Two video lists are merged by path into `allVideos`:

- `discoveredVideos` — what the scan/stream returned.
- `downloadedVideos` — optimistic rows (a file the UI knows about before the next full scan replaces it).

If the same path is in both, the discovered row wins. Uploads and moves usually patch `discoveredVideos` directly via `upsertDiscoveredVideo` / `patchVideoPath` when the API returns a `video` object.

`filteredVideos` is `allVideos` after:

- folder (`__all__` or an exact folder string),
- favorites only,
- every selected tag (AND),
- substring search on name and folder,
- optional `activeBatchId`,

then sorted. Sort keys: name, name-desc, folder, ext, modified, modified-desc, size, size-desc. The choice is saved in `localStorage` under `vidvault-sort`.

Duplicate groups are computed in the client: group by hash, keep groups with more than one member, and pick a `keepPath` (newer mtime, then larger size). The duplicate banner can auto-open the duplicate modal once per load when groups exist.

### Loading

Desktop `init()`:

1. In parallel: folders, tags, collections.
2. In parallel: batches, and `GET /api/videos/stream`.
3. Incoming video events are buffered and flushed into the store every 200 ms.
4. On stream failure, fall back to `GET /api/videos`.
5. Recompute duplicates and banner visibility. The first desktop load may auto-open the duplicate modal.

`refresh({ rescan })` reloads folders, tags, collections, and batches, clears the video list, and streams again with `?rescan=1` by default.

Mobile (viewport max-width 600px, or a phone user agent) does not stream the whole library. It pages `GET /api/videos?offset&limit`: 5 videos first, then 25 per “Load more.” Duplicate auto-open is suppressed because the list is partial. Filters and counts only see what has been paged in.

### Gallery

`Gallery.svelte` renders `filteredVideos`. A non-empty list goes through `VirtualGallery`, which uses `@tanstack/svelte-virtual`. It chunks videos into rows from the container width (minimum column about 260px, 12px gap, 16:9 thumb plus a fixed meta block). List view is one video per row. Only rows near the scrollport, plus a small overscan, mount `VideoCard`s. The scroll parent is the `<main>` element.

Each card can preview on hover, open the lightbox, toggle selection, and open the context menu (move, tag, delete, rename, collection, and similar actions). Dragging a card onto a sidebar folder calls the move API.

When a batch filter is active, the gallery shows a “Mark as resolved” button that clears `is_new` for the paths in the current filtered list.

### Sidebar

Top to bottom: Discover, batches, tags, collections, folders, preview settings. On a narrow layout the sidebar is an overlay toggled from the header (`sidebarOpen`).

### Selection

Select mode plus `selectedPaths` drives `SelectBar`: move many, tag many, add many to a collection, delete many. Bulk operations loop the single-item APIs (or the bulk tag/collection endpoints where those exist) and patch local state as each call returns.

---

## 14. Request path, end to end

Example: open the app, then move one file into a folder.

1. Browser loads `GET /`. The Go server returns embedded `index.html`. Scripts come from `GET /assets/…`.
2. `bootstrap()` calls `GET /api/auth/status`. No PIN, so it calls `init()`.
3. Parallel fetches fill folders, tags, collections, and batches.
4. `GET /api/videos/stream` hits `handleVideosStream`. The startup scan may already have set `library.ready`. If so, the handler does not walk the disk; it writes one NDJSON line per video. If not, this request (or the background scan) walks, hashes, and runs `MarkVideos`.
5. `MarkVideos` sees the root in `baselined_roots`. Hashes already in `seen_files.json` keep their reviewed flag. A hash that is absent becomes `is_new: true` and joins the current discovered batch.
6. The client flushes videos into `discoveredVideos`. `filteredVideos` updates. `VirtualGallery` paints rows.
7. User drops a card on folder `trips`. `POST /api/move` renames the file on disk, updates `VideoLibrary`, returns the new video object.
8. `patchVideoPath` replaces the old path in the store. Folder counts recompute. The card shows the new folder. Tags stay because the hash did not change.
9. Clicking play sets `playingPath`. The modal’s `<video src="/video?path=trips/clip.mp4">` sends a range request. `ServeContent` returns 206 with the requested bytes.

Example: first run against a new media root.

1. `seen_files.json` has no baseline for that absolute path.
2. Startup `MarkVideos` writes every current hash with `reviewed_at` set, and sets `baselined_roots`.
3. Nothing is “new.” The incoming banner stays hidden.
4. You copy ten new mp4s into the folder and refresh with a rescan.
5. Those ten hashes are missing, so they are new, share one discovered batch if they arrive inside the same 30 minute window, and the banner shows that batch’s `new_count`.

---

## 15. Logging

`src/logging/logger.go` is slog with a small colorized writer and a per-file logger name (`main`, `server`, `video`, `folders`, `favorites`, `tags`, `collections`, `template`, `browser`). Every HTTP request logs method, path, status, and duration. Debug mode adds bytes, remote address, user agent, and the noisy scan/hash traces. `-d` (or `VIDVAULT_DEBUG=1` in Docker) turns that on.

---

## 16. What is intentionally simple

- No database. Five JSON files and an in-memory map.
- No users. Optional shared PIN, one cookie secret per process.
- No transcoding, thumbnails, or subtitle index. The browser plays the file or it does not.
- No watcher on the media root. New files copied in by another program appear on the next scan (`?rescan=1` or a process restart’s startup scan).
- Deletes do not cascade into tags, favorites, or collections. Orphan hashes sit in those files until you remove them there.
- The scan cache and library vanish when the process exits.
- Mobile never holds the full library in memory; desktop does, and virtualizes only the DOM.

---

## 17. File index

| Path | Role |
| --- | --- |
| `src/main.go` | Flags, listen, print URLs, open browser |
| `src/server.go` | Routes and handlers |
| `src/auth.go` | PIN cookie |
| `src/browser.go` | LAN address and OS browser |
| `src/template.go` | Embed `web/dist` |
| `src/video/video.go` | Extensions, hash, scan, rename, delete |
| `src/video/library.go` | In-memory path index |
| `src/video/progress.go` | Terminal scan progress |
| `src/folders.go` | Folder walk, mkdir, rmdir, move |
| `src/favorites.go` | `favorites.json` |
| `src/tags.go` | `tags.json` |
| `src/collections.go` | `watch_collections.json` |
| `src/seen_files.go` | `seen_files.json` |
| `src/upload_batches.go` | `upload_batches.json` |
| `src/logging/logger.go` | slog setup |
| `frontend/src/App.svelte` | Shell |
| `frontend/src/lib/api.ts` | HTTP client |
| `frontend/src/lib/stores/library.ts` | Library, filters, load |
| `frontend/src/lib/stores/ui.ts` | Modals and preferences |
| `frontend/src/components/gallery/` | Gallery, virtual rows, cards, discover, batch header |
| `frontend/src/components/sidebar/` | Batches, tags, collections, folders |
| `frontend/src/components/modals/` | Player, upload, move, tags, duplicates, incoming, collections, preview settings |
| `frontend/vite.config.ts` | Dev proxy and build output `../src/web/dist` |
| `scripts/generate-test-videos.sh` | ffmpeg seed data for `./data` |
| `notes/DESIGN.md` | Visual language (some paths in that note still say `src/web/styles.css`; colors now live in `frontend/src/app.css`) |
| `notes/NOTES.md` | Docker media-root warning and logging notes |
