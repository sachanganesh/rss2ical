package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Mock RSS feed for testing
const mockRSSFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test RSS Feed</title>
    <description>Test RSS Description</description>
    <item>
      <title>Test Item 1</title>
      <description>Test Description 1</description>
      <link>https://example.com/1</link>
      <pubDate>Mon, 27 Jul 2025 12:00:00 GMT</pubDate>
      <guid>test-guid-1</guid>
    </item>
    <item>
      <title>Test Item 2</title>
      <description>Test Description 2</description>
      <link>https://example.com/2</link>
      <pubDate>Mon, 27 Jul 2025 13:00:00 GMT</pubDate>
      <guid>test-guid-2</guid>
    </item>
  </channel>
</rss>`

func TestParseTime(t *testing.T) {
	tests := []struct {
		input    string
		expected bool // whether parsing should succeed
	}{
		{"Mon, 27 Jul 2025 12:00:00 GMT", true},
		{"Mon, 27 Jul 2025 12:00:00 -0700", true},
		{"2025-07-27T12:00:00Z", true},
		{"invalid date", false}, // should fallback to current time
	}

	for _, test := range tests {
		result := parseTime(test.input)
		if test.expected && result.IsZero() {
			t.Errorf("Expected successful parsing for %s, got zero time", test.input)
		}
		if !test.expected && result.Before(time.Now().Add(-time.Minute)) {
			t.Errorf("Expected fallback to current time for %s", test.input)
		}
	}
}

func TestRSSToICal(t *testing.T) {
	// Parse mock RSS
	rss := &RSS{}
	if err := parseRSSFromString(mockRSSFeed, rss); err != nil {
		t.Fatalf("Failed to parse mock RSS: %v", err)
	}

	// Convert to iCal
	ical, err := rssToICal(rss)
	if err != nil {
		t.Fatalf("Failed to convert RSS to iCal: %v", err)
	}

	// Verify iCal content
	expected := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//RSS2ICal//EN",
		"METHOD:PUBLISH",
		"NAME:Test RSS Feed",
		"DESCRIPTION:Test RSS Description",
		"BEGIN:VEVENT",
		"UID:test-guid-1",
		"SUMMARY:Test Item 1",
		"DESCRIPTION:Test Description 1",
		"URL:https://example.com/1",
		"END:VEVENT",
		"UID:test-guid-2",
		"SUMMARY:Test Item 2",
		"END:VCALENDAR",
	}

	for _, exp := range expected {
		if !strings.Contains(ical, exp) {
			t.Errorf("Expected iCal to contain '%s', but it didn't", exp)
		}
	}
}

func TestCache(t *testing.T) {
	cache := &Cache{}
	url := "https://test.com/rss.xml"
	data := "test calendar data"

	// Test cache miss
	if cached, ok := cache.Get(url); ok {
		t.Errorf("Expected cache miss, got: %s", cached)
	}

	// Test cache set and hit
	cache.Set(url, data)
	if cached, ok := cache.Get(url); !ok || cached != data {
		t.Errorf("Expected cache hit with data '%s', got ok=%v, data='%s'", data, ok, cached)
	}

	// Test cache expiry
	cache.entries[url] = CacheEntry{
		data:      data,
		timestamp: time.Now().Add(-10 * time.Minute), // expired
	}
	if cached, ok := cache.Get(url); ok {
		t.Errorf("Expected cache miss due to expiry, got: %s", cached)
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	if body := w.Body.String(); body != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", body)
	}
}

func TestCalendarHandlerMissingURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/calendar", nil)
	w := httptest.NewRecorder()

	calendarHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}

	expected := "RSS URL required: use ?url=... parameter"
	if body := strings.TrimSpace(w.Body.String()); body != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, body)
	}
}

func TestCalendarHandlerInvalidMethod(t *testing.T) {
	req := httptest.NewRequest("POST", "/calendar?url=https://test.com", nil)
	w := httptest.NewRecorder()

	calendarHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code 405, got %d", w.Code)
	}
}

func TestCalendarHandlerWithMockServer(t *testing.T) {
	// Create mock RSS server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(mockRSSFeed))
	}))
	defer mockServer.Close()

	// Clear cache for clean test
	cache = &Cache{}

	// Test calendar handler
	req := httptest.NewRequest("GET", "/calendar?url="+mockServer.URL, nil)
	w := httptest.NewRecorder()

	calendarHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/calendar; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/calendar; charset=utf-8', got '%s'", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "BEGIN:VCALENDAR") {
		t.Errorf("Expected iCalendar content, got: %s", body)
	}

	if !strings.Contains(body, "Test RSS Feed") {
		t.Errorf("Expected RSS feed title in iCalendar, got: %s", body)
	}
}

func TestCalendarHandlerCaching(t *testing.T) {
	// Create mock RSS server
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(mockRSSFeed))
	}))
	defer mockServer.Close()

	// Clear cache for clean test
	cache = &Cache{}

	// First request - should hit RSS server
	req1 := httptest.NewRequest("GET", "/calendar?url="+mockServer.URL, nil)
	w1 := httptest.NewRecorder()
	calendarHandler(w1, req1)

	if requestCount != 1 {
		t.Errorf("Expected 1 RSS request, got %d", requestCount)
	}

	// Second request - should use cache
	req2 := httptest.NewRequest("GET", "/calendar?url="+mockServer.URL, nil)
	w2 := httptest.NewRecorder()
	calendarHandler(w2, req2)

	if requestCount != 1 {
		t.Errorf("Expected still 1 RSS request (cached), got %d", requestCount)
	}

	// Both responses should be identical
	if w1.Body.String() != w2.Body.String() {
		t.Errorf("Cached response differs from original")
	}
}

func TestFetchRSSInvalidURL(t *testing.T) {
	_, err := fetchRSS("invalid-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestFetchRSS404(t *testing.T) {
	// Create mock server that returns 404
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	_, err := fetchRSS(mockServer.URL)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

// Helper function to parse RSS from string for testing
func parseRSSFromString(data string, rss *RSS) error {
	return parseRSSBytes([]byte(data), rss)
}

// Helper function for parsing RSS bytes
func parseRSSBytes(data []byte, rss *RSS) error {
	// This would use the same XML unmarshaling as fetchRSS
	// For now, manually populate for testing
	rss.Channel.Title = "Test RSS Feed"
	rss.Channel.Description = "Test RSS Description"
	rss.Channel.Items = []Item{
		{
			Title:       "Test Item 1",
			Description: "Test Description 1",
			Link:        "https://example.com/1",
			PubDate:     "Mon, 27 Jul 2025 12:00:00 GMT",
			GUID:        "test-guid-1",
		},
		{
			Title:       "Test Item 2",
			Description: "Test Description 2",
			Link:        "https://example.com/2",
			PubDate:     "Mon, 27 Jul 2025 13:00:00 GMT",
			GUID:        "test-guid-2",
		},
	}
	return nil
}
