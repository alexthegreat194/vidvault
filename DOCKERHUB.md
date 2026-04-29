# Vidvault (Docker)

Local **video gallery** HTTP server in Go: mount a media folder, browse and stream in the browser.

## Security

**No built-in authentication.** Anyone who can reach the server can list and stream files under the media root. Use on **trusted networks** or restrict access (for example `localhost` port mapping and a firewall).

## Quick start

Replace the image name with your tag if it differs (for example `alexthegreat194/vidvault:latest`).

```bash
docker run --rm -p 8765:8765 -v /path/to/your/videos:/data alexthegreat194/vidvault:latest -p 8765 /data
```

Then open `http://localhost:8765/`.

## Image details

- **Multi-stage build:** static Linux binary (`CGO_ENABLED=0`), runtime on **Alpine** with `ca-certificates`.
- **WORKDIR:** `/data` (intended media root).
- **Listen port (inside container):** `8765`.

## Volumes

| Host path      | Container path | Purpose       |
| -------------- | -------------- | ------------- |
| `/path/to/lib` | `/data`        | Media library |

## Ports

| Container | Example host mapping |
| --------- | -------------------- |
| `8765`    | `-p 8765:8765`       |

Use `-p HOST:8765` to map a different host port.

## Command line

Flags use Go’s `flag` package: **put all `-flags` before** the optional directory. Default command is `-p 8765 /data`.

```bash
docker run --rm -p 9000:9000 -v ~/Videos:/data alexthegreat194/vidvault:latest -p 9000 /data
```

## Environment variables

The entrypoint maps these to CLI flags. Truthy values: `1`, `true`, `yes`, `on` (case-insensitive).

| Variable                     | Maps to            |
| ---------------------------- | ------------------ |
| `VIDVAULT_DEBUG`             | `-d`               |
| `VIDVAULT_DISABLE_BROWSER` | `-disable-browser` |
| `VIDVAULT_PIN`               | `-pin` (value)     |

```bash
docker run --rm -p 8765:8765 -v /path/to/videos:/data \
  -e VIDVAULT_DISABLE_BROWSER=1 \
  alexthegreat194/vidvault:latest -p 8765 /data
```

## Supported video extensions

`.webm` `.mp4` `.mpv` `.mkv` `.mov` `.avi` `.m4v` `.ogv` (case-insensitive)

## Links

- **Source:** https://github.com/alexthegreat194/vidvault  
- **Issues:** https://github.com/alexthegreat194/vidvault/issues  
