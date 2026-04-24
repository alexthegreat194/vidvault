# Docker notes

## Media root must not be the container `/`

The server uses one directory as the library: it walks that tree for folders and videos, and stores uploads there. If you run the image with `.` and no `WORKDIR`, the process starts in `/`, so the app treats **the whole container filesystem** as the library. The sidebar will list real paths like `bin`, `dev`, `etc` (Alpine layout), which looks like “random” folders.

**Use a single data directory** (e.g. `/data` in the image) and **bind-mount** your host videos there when you run the container.

## Build and run

- **Image default:** the Dockerfile sets `WORKDIR` to `/data` and the default command serves `/data` (empty until you mount something).
- **Convenience script:** from the repo root, run `./scripts/docker-run.sh` (builds the image and runs with a host folder mounted at `/data`).

### Environment variables (script)

| Variable   | Default        | Meaning                          |
| ---------- | -------------- | -------------------------------- |
| `IMAGE`    | `vidvault:local` | Docker image tag after build   |
| `PORT`     | `8765`         | Host port → container 8765       |
| `DATA_DIR` | `<repo>/data`  | Host directory mounted as `/data` |

Example:

```bash
PORT=9000 DATA_DIR="$HOME/Movies" ./scripts/docker-run.sh
```

Then open `http://localhost:9000` (or the chosen port).

### Manual `docker run`

```bash
docker build -t vidvault:local .
mkdir -p ./data
docker run --rm -p 8765:8765 -v "$(pwd)/data:/data" vidvault:local -p 8765 /data
```

## Thumbnails

Grid “thumbnails” are **inline `<video>` previews** (not separate image files). If a file fails to load or the browser cannot decode the codec, tiles may stay dark; fixing the library path and permissions is the first step.
