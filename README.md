# RSS2ICal

Lightweight Go service that converts RSS feeds to iCalendar format for consumption by Google Calendar, Apple Calendar, etc.

## Usage

### Web Interface (Recommended)

Visit the home page in your browser to easily generate calendar URLs:

```
http://localhost:8080/
```

1. Paste any RSS feed URL 
2. Click "Generate Calendar URL"
3. Copy the generated URL
4. Add to your calendar app

### Direct API Usage

```bash
# Start the server
go run main.go

# Convert any RSS feed on-demand (URL must be properly encoded)
curl "http://localhost:8080/calendar?url=https%3A//feeds.bbci.co.uk/news/rss.xml"

# For RSS URLs with query parameters, use the web interface or encode manually
```

## Endpoints

- `GET /` - Home page with URL generation form
- `GET /calendar?url=<ENCODED_RSS_URL>` - Converts RSS feed to iCalendar format
- `GET /health` - Health check

## Environment Variables

- `PORT` - Server port (default: 8080)

## Features

- **Web Interface**: Simple form to generate properly encoded calendar URLs
- **Dynamic RSS URLs**: Support any RSS feed via query parameter
- **Automatic URL Encoding**: JavaScript handles complex URLs with parameters
- **Per-URL Caching**: 5-minute TTL for fast responses
- **Concurrent-Safe**: Thread-safe cache operations
- **Date Format Handling**: Supports common RSS date formats
- **Calendar App Ready**: Proper HTTP headers for Google Calendar, Apple Calendar, etc.
- **Copy-to-Clipboard**: One-click URL copying from web interface

## Example RSS Feeds

The web interface includes examples like:
- SF Recreation & Parks volunteer events
- Any RSS 2.0 compatible feed 

## Testing

```bash
# Run tests
go test

# Run tests with coverage
go test -cover

# Verbose test output
go test -v
```

Test coverage: 85%+ of statements

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