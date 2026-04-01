# Stockyard Outpost

**Personal dashboard — weather, RSS feeds, bookmarks, habit tracker, new tab replacement**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted developer tools.

## Quick Start

```bash
docker run -p 9220:9220 -v outpost_data:/data ghcr.io/stockyard-dev/stockyard-outpost
```

Or with docker-compose:

```bash
docker-compose up -d
```

Open `http://localhost:9220` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9220` | HTTP port |
| `DATA_DIR` | `./data` | SQLite database directory |
| `OUTPOST_LICENSE_KEY` | *(empty)* | Pro license key |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 5 widgets, 3 habits | Unlimited widgets and habits |
| Price | Free | $1.99/mo |

Get a Pro license at [stockyard.dev/tools/](https://stockyard.dev/tools/).

## Category

Operations & Teams

## License

Apache 2.0
