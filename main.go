package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	ics "github.com/arran4/golang-ical"
)

const (
	defaultPort = "8080"
	cacheTTL    = 5 * time.Minute
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

type CacheEntry struct {
	data      string
	timestamp time.Time
}

type Cache struct {
	entries map[string]CacheEntry
	mu      sync.RWMutex
}

func (c *Cache) Get(url string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[url]
	if !exists || time.Since(entry.timestamp) > cacheTTL {
		return "", false
	}
	return entry.data, true
}

func (c *Cache) Set(url, data string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[string]CacheEntry)
	}
	c.entries[url] = CacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
}

var cache = &Cache{}

func fetchRSS(url string) (*RSS, error) {
	log.Printf("Fetching RSS from: %s", url)

	// Create request with proper headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to mimic a real browser
	req.Header.Set("User-Agent", "RSS2ICal/1.0 (Go HTTP Client)")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, */*")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP GET error: %v", err)
		return nil, fmt.Errorf("failed to fetch RSS: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("RSS fetch status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS fetch returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS body: %w", err)
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("failed to parse RSS: %w", err)
	}

	return &rss, nil
}

func parseTime(pubDate string) time.Time {
	// Try common RSS date formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, pubDate); err == nil {
			return t
		}
	}

	// Fallback to current time if parsing fails
	return time.Now()
}

func rssToICal(rss *RSS) (string, error) {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//RSS2ICal//EN")
	cal.SetName(rss.Channel.Title)
	cal.SetDescription(rss.Channel.Description)

	for _, item := range rss.Channel.Items {
		event := cal.AddEvent(item.GUID)
		event.SetSummary(item.Title)
		event.SetDescription(item.Description)
		event.SetURL(item.Link)

		startTime := parseTime(item.PubDate)
		event.SetStartAt(startTime)
		event.SetEndAt(startTime.Add(time.Hour)) // Default 1-hour duration

		event.SetCreatedTime(startTime)
		event.SetModifiedAt(startTime)
	}

	return cal.Serialize(), nil
}

func calendarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get RSS URL from query parameter
	rssURL := r.URL.Query().Get("url")
	if rssURL == "" {
		http.Error(w, "RSS URL required: use ?url=... parameter", http.StatusBadRequest)
		return
	}

	// Check cache first
	if cached, ok := cache.Get(rssURL); ok {
		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cached))
		return
	}

	// Fetch fresh data
	rss, err := fetchRSS(rssURL)
	if err != nil {
		log.Printf("Error fetching RSS from %s: %v", rssURL, err)
		http.Error(w, "Failed to fetch RSS feed", http.StatusInternalServerError)
		return
	}

	ical, err := rssToICal(rss)
	if err != nil {
		log.Printf("Error converting to iCal: %v", err)
		http.Error(w, "Failed to convert to iCalendar", http.StatusInternalServerError)
		return
	}

	// Cache the result
	cache.Set(rssURL, ical)

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ical))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/calendar", calendarHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting RSS2ICal server on port %s", port)
	log.Printf("Calendar endpoint: http://localhost:%s/calendar?url=<RSS_URL>", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
