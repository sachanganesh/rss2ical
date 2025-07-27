# RSS2ICal

Lightweight Go service that converts RSS feeds to iCalendar format for consumption by Google Calendar, Apple Calendar, etc.

## Usage

```bash
# Install dependencies
go mod download

# Run with default settings
go run main.go

# Or set your RSS URL
RSS_URL="https://your-rss-feed.com/feed.xml" go run main.go

# Custom port
PORT=3000 RSS_URL="https://example.com/rss.xml" go run main.go
```

## Endpoints

- `GET /calendar` - Returns iCalendar data
- `GET /health` - Health check

## Environment Variables

- `RSS_URL` - RSS feed URL to convert
- `PORT` - Server port (default: 8080)

## Features

- In-memory caching (5min TTL)
- Concurrent-safe
- Handles common RSS date formats
- Proper HTTP headers for calendar apps 