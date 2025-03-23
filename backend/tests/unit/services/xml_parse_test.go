package services

import (
	"encoding/xml"
	"testing"
	"time"
)

func TextNotificationParse(t *testing.T) {
	// Sample XML from res.xml
	sampleXML := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns:yt="http://www.youtube.com/xml/schemas/2015" xmlns="http://www.w3.org/2005/Atom">
  <link rel="hub" href="https://pubsubhubbub.appspot.com"/>
  <link rel="self" href="https://www.youtube.com/xml/feeds/videos.xml?channel_id=UCVr7u3eUqwtt6HMtLstE8Aw"/>
  <title>YouTube video feed</title>
  <updated>2025-03-20T13:58:33.182347455+00:00</updated>
  <entry>
    <id>yt:video:E9QpwCVPPyM</id>
    <yt:videoId>E9QpwCVPPyM</yt:videoId>
    <yt:channelId>UCVr7u3eUqwtt6HMtLstE8Aw</yt:channelId>
    <title>Screen Recording 2025 03 18 at 01 40 04</title>
    <link rel="alternate" href="https://www.youtube.com/watch?v=E9QpwCVPPyM"/>
    <author>
      <n>waxen</n>
      <uri>https://www.youtube.com/channel/UCVr7u3eUqwtt6HMtLstE8Aw</uri>
    </author>
    <published>2025-03-20T13:57:41+00:00</published>
    <updated>2025-03-20T13:58:33.182347455+00:00</updated>
  </entry>
</feed>`

	var feed Feed
	err := xml.Unmarshal([]byte(sampleXML), &feed)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if len(feed.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(feed.Entries))
	}

	entry := feed.Entries[0]
	if entry.ID != "yt:video:E9QpwCVPPyM" {
		t.Errorf("Expected ID 'yt:video:E9QpwCVPPyM', got '%s'", entry.ID)
	}
	if entry.VideoID != "E9QpwCVPPyM" {
		t.Errorf("Expected VideoID 'E9QpwCVPPyM', got '%s'", entry.VideoID)
	}
	if entry.ChannelID != "UCVr7u3eUqwtt6HMtLstE8Aw" {
		t.Errorf("Expected ChannelID 'UCVr7u3eUqwtt6HMtLstE8Aw', got '%s'", entry.ChannelID)
	}
	if entry.Title != "Screen Recording 2025 03 18 at 01 40 04" {
		t.Errorf("Expected Title 'Screen Recording 2025 03 18 at 01 40 04', got '%s'", entry.Title)
	}
	if entry.Link.Rel != "alternate" {
		t.Errorf("Expected Link.Rel 'alternate', got '%s'", entry.Link.Rel)
	}
	if entry.Link.Href != "https://www.youtube.com/watch?v=E9QpwCVPPyM" {
		t.Errorf("Expected Link.Href 'https://www.youtube.com/watch?v=E9QpwCVPPyM', got '%s'", entry.Link.Href)
	}
	if entry.Author.Name != "waxen" {
		t.Errorf("Expected Author.Name 'waxen', got '%s'", entry.Author.Name)
	}
	if entry.Author.URI != "https://www.youtube.com/channel/UCVr7u3eUqwtt6HMtLstE8Aw" {
		t.Errorf("Expected Author.URI 'https://www.youtube.com/channel/UCVr7u3eUqwtt6HMtLstE8Aw', got '%s'", entry.Author.URI)
	}

	// Test date parsing
	expectedPublished := time.Date(2025, 3, 20, 13, 57, 41, 0, time.UTC)
	publishedAt, err := time.Parse(time.RFC3339, entry.Published)
	if err != nil {
		t.Fatalf("Failed to parse published date: %v", err)
	}
	if !publishedAt.Equal(expectedPublished) {
		t.Errorf("Expected published date %v, got %v", expectedPublished, publishedAt)
	}
} 