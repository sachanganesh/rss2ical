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

## Deployment

### Production Deployment (Recommended)

Uses pre-built images from GitHub Actions - fast and lightweight:

```bash
# Deploy with pre-built image
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f rss2ical

# Update to latest
docker-compose -f docker-compose.prod.yml pull && docker-compose -f docker-compose.prod.yml up -d

# Stop
docker-compose -f docker-compose.prod.yml down
```

**Default port**: 8081 (external) → 8080 (internal)

### Port Configuration

To change the external port, edit `docker-compose.prod.yml`:
```yaml
ports:
  - "9000:8080"  # Change 8081 → 9000, keep internal 8080
```

For nginx reverse proxy setup:
```nginx
# /etc/nginx/sites-available/rss.yourdomain.com
server {
    server_name rss.yourdomain.com;
    location / {
        proxy_pass http://localhost:8081;  # Match your external port
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Development Deployment

For local development with live building:

```bash
# Build and run locally
docker-compose up -d

# View logs
docker-compose logs -f rss2ical
```

### CI/CD

GitHub Actions automatically builds and pushes images on every commit to main:
- **Image**: `ghcr.io/sachanganesh/rss2ical:latest`
- **Build time**: ~2-3 minutes (vs 4+ minutes on small droplets)
- **Free**: For public repositories

**Image size**: ~15MB Alpine-based container  
**Memory usage**: ~16-64MB with limits  
**Auto-restart**: Enabled for production reliability 