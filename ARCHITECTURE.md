# RSS2ICal

Provide an on-demand source of truth for converting any RSS feed to iCalendar format for consumption by GCal or Apple Calendar.

## Goals
- Provide a stable endpoint that converts any RSS feed to iCalendar format
- Support dynamic RSS URLs via query parameters
- Ensure up-to-date calendar data is always available on-demand
- Support integration with major calendar applications (Google Calendar, Apple Calendar, etc.)
- Minimize server load and response time through intelligent caching

## Non-Goals
- Active push notifications to calendar applications
- User authentication or personalized feeds
- Complex event processing or modification
- Persistent storage beyond in-memory caching

## Dependencies
- Go programming language (version 1.16+)
- External libraries:
  - github.com/arran4/golang-ical (for iCalendar generation)
  - encoding/xml (standard library, for RSS parsing)
- Web server capable of running Go applications

## Architecture

### Components

1. HTTP Server
   - Listens for incoming GET requests
   - Handles routing to appropriate handlers
   - Extracts RSS URLs from query parameters

2. RSS Fetcher
   - On-demand fetches RSS feeds from user-provided URLs
   - Implements error handling and timeout logic
   - Supports any valid RSS 2.0 feed

3. RSS Parser
   - Parses the XML content of RSS feeds
   - Extracts relevant event information (title, description, date, link)
   - Handles various RSS date formats

4. iCalendar Generator
   - Converts parsed RSS data into iCalendar format
   - Utilizes the golang-ical library
   - Creates properly formatted VEVENT entries

5. Cache Manager
   - Stores generated iCalendar data per RSS URL
   - Implements TTL-based cache invalidation (5 minutes)
   - Thread-safe operations with read/write mutexes

6. Request Handler
   - Serves cached or freshly generated iCalendar data
   - Implements appropriate HTTP headers for calendar applications
   - Validates RSS URL parameters

### Data Flow

1. Client requests `/calendar?url=<RSS_URL>`
2. Request Handler validates the required RSS URL parameter
3. Cache Manager checks for cached iCalendar data for this URL
4. If cached and fresh (< 5min old), serve cached data immediately
5. If not cached or stale:
   - RSS Fetcher retrieves the RSS feed from the specified URL
   - RSS Parser extracts event information from the feed
   - iCalendar Generator converts parsed data to iCalendar format
   - Cache Manager stores the generated data with timestamp
6. Request Handler serves the iCalendar data with appropriate headers

### Key Considerations

- **Per-URL Caching**: Each RSS URL has its own cache entry with independent TTL
- **Concurrent Safety**: All cache operations are protected by read/write mutexes
- **Error Handling**: Robust error handling for RSS fetching, parsing, and malformed URLs
- **Performance**: Sub-15ms response times for cached content
- **Memory Usage**: In-memory cache with automatic TTL-based cleanup
- **Rate Limiting**: Natural rate limiting through caching reduces upstream RSS server load

## API

### Endpoint: `/calendar`
- **Method**: GET
- **Parameters**: 
  - `url` (required): RSS feed URL to convert
- **Response**: 
  - Content-Type: `text/calendar; charset=utf-8`
  - Cache-Control: `public, max-age=300`
  - Body: iCalendar formatted data (RFC 5545)

### Example Usage
```
GET /calendar?url=https://feeds.bbci.co.uk/news/rss.xml
```

### Endpoint: `/health`
- **Method**: GET
- **Response**: 
  - Status: 200 OK
  - Body: "OK"

## Scalability

The current architecture supports unlimited RSS feeds through dynamic URLs:
- **Horizontal Scaling**: Stateless design allows multiple instances behind a load balancer
- **Containerization**: Ready for Docker deployment and orchestration
- **Memory Management**: Cache naturally limits memory usage through TTL expiration
- **Load Distribution**: Each RSS URL is cached independently, distributing load
- **Future Enhancements**: Could add Redis for shared caching across instances

## Security Considerations

- Ensure the server is properly configured and kept updated.
- Implement HTTPS to encrypt data in transit.
- Consider implementing request validation to prevent potential abuse.

## Monitoring and Maintenance

- Implement logging for error tracking and usage statistics.
- Set up alerts for critical errors (e.g., persistent failures in RSS fetching).
- Regularly review and update dependencies to address potential security vulnerabilities.
