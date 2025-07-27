# RSS2ICal

Lightweight Go service that converts RSS feeds to iCalendar format for consumption by Google Calendar, Apple Calendar, etc.

## Usage

```bash
# Start the server
go run main.go

# Convert any RSS feed on-demand
curl "http://localhost:8080/calendar?url=https://feeds.bbci.co.uk/news/rss.xml"

# Use with calendar apps - just add this URL:
# http://your-server:8080/calendar?url=https://your-rss-feed.com/feed.xml
```

## Endpoints

- `GET /calendar?url=<RSS_URL>` - Converts RSS feed to iCalendar format
- `GET /health` - Health check

## Environment Variables

- `PORT` - Server port (default: 8080)

## Features

- Dynamic RSS URL via query parameter
- Per-URL caching (5min TTL)
- Concurrent-safe
- Handles common RSS date formats  
- Proper HTTP headers for calendar apps

## Testing

```bash
# Run tests
go test

# Run tests with coverage
go test -cover

# Verbose test output
go test -v
```

Test coverage: 77.4% of statements

## Docker Deployment

Perfect for multi-app servers (alongside LibreChat, etc.):

```bash
# Build and run
docker-compose up -d

# View logs
docker-compose logs -f rss2ical

# Update
docker-compose pull && docker-compose up -d

# Stop
docker-compose down
```

**Image size**: ~15MB Alpine-based container  
**Memory usage**: ~16-64MB with limits  
**Auto-restart**: Enabled for production reliability 