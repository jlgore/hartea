package har

import (
	"sort"
	"strings"
	"time"
)

type Metrics struct {
	TotalRequests     int
	TotalTime         float64
	TotalSize         int64
	TTFB              float64
	PageLoadTime      float64
	DNSTime           float64
	ConnectTime       float64
	SSLTime           float64
	FirstContentfulPaint float64
	LargestContentfulPaint float64
	CacheHitRatio     float64
	ThirdPartyRequests int
	ErrorRequests     int
}

type Analyzer struct {
	har *HAR
}

func NewAnalyzer(har *HAR) *Analyzer {
	return &Analyzer{har: har}
}

func (a *Analyzer) CalculateMetrics() *Metrics {
	entries := a.har.Log.Entries
	if len(entries) == 0 {
		return &Metrics{}
	}

	metrics := &Metrics{
		TotalRequests: len(entries),
	}

	var totalSize int64
	var totalTime float64
	var dnsTime, connectTime, sslTime float64
	var cacheHits int
	var errorRequests int
	var thirdPartyRequests int
	var firstByte float64 = -1

	// Get page load time from page timings if available
	if len(a.har.Log.Pages) > 0 {
		page := a.har.Log.Pages[0]
		if page.PageTimings.OnLoad > 0 {
			metrics.PageLoadTime = float64(page.PageTimings.OnLoad)
		}
	}

	for _, entry := range entries {
		// Total time and size
		totalTime += entry.Time
		totalSize += int64(entry.Response.Content.Size)

		// Response status analysis
		if entry.Response.Status >= 400 {
			errorRequests++
		}

		// Timing analysis
		if entry.Timings.DNS > 0 {
			dnsTime += float64(entry.Timings.DNS)
		}
		if entry.Timings.Connect > 0 {
			connectTime += float64(entry.Timings.Connect)
		}
		if entry.Timings.SSL > 0 {
			sslTime += float64(entry.Timings.SSL)
		}

		// TTFB calculation (first request wait time)
		if firstByte == -1 || (entry.Timings.Wait > 0 && float64(entry.Timings.Wait) < firstByte) {
			firstByte = float64(entry.Timings.Wait)
		}

		// Cache analysis
		if entry.Cache.BeforeRequest != nil {
			cacheHits++
		}

		// Third-party analysis
		if a.isThirdParty(entry.Request.URL) {
			thirdPartyRequests++
		}
	}

	metrics.TotalTime = totalTime
	metrics.TotalSize = totalSize
	metrics.TTFB = firstByte
	metrics.DNSTime = dnsTime / float64(len(entries))
	metrics.ConnectTime = connectTime / float64(len(entries))
	metrics.SSLTime = sslTime / float64(len(entries))
	metrics.CacheHitRatio = float64(cacheHits) / float64(len(entries)) * 100
	metrics.ThirdPartyRequests = thirdPartyRequests
	metrics.ErrorRequests = errorRequests

	// If no page load time from page timings, estimate from entries
	if metrics.PageLoadTime == 0 {
		metrics.PageLoadTime = a.calculateEstimatedPageLoadTime()
	}

	return metrics
}

func (a *Analyzer) GetSlowestRequests(limit int) []Entry {
	entries := make([]Entry, len(a.har.Log.Entries))
	copy(entries, a.har.Log.Entries)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time > entries[j].Time
	})

	if limit > len(entries) {
		limit = len(entries)
	}

	return entries[:limit]
}

func (a *Analyzer) GetLargestRequests(limit int) []Entry {
	entries := make([]Entry, len(a.har.Log.Entries))
	copy(entries, a.har.Log.Entries)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Response.Content.Size > entries[j].Response.Content.Size
	})

	if limit > len(entries) {
		limit = len(entries)
	}

	return entries[:limit]
}

func (a *Analyzer) GetErrorRequests() []Entry {
	var errors []Entry
	for _, entry := range a.har.Log.Entries {
		if entry.Response.Status >= 400 {
			errors = append(errors, entry)
		}
	}
	return errors
}

func (a *Analyzer) GetResourcesByType() map[string][]Entry {
	resources := make(map[string][]Entry)
	
	for _, entry := range a.har.Log.Entries {
		contentType := entry.Response.Content.MimeType
		if contentType == "" {
			contentType = "unknown"
		}
		
		// Simplify content types
		if strings.Contains(contentType, "javascript") {
			contentType = "javascript"
		} else if strings.Contains(contentType, "css") {
			contentType = "css"
		} else if strings.Contains(contentType, "image") {
			contentType = "image"
		} else if strings.Contains(contentType, "html") {
			contentType = "html"
		} else if strings.Contains(contentType, "json") {
			contentType = "json"
		} else if strings.Contains(contentType, "font") {
			contentType = "font"
		}
		
		resources[contentType] = append(resources[contentType], entry)
	}
	
	return resources
}

func (a *Analyzer) isThirdParty(url string) bool {
	// Simple third-party detection based on common patterns
	thirdPartyDomains := []string{
		"googleapis.com",
		"googletagmanager.com",
		"facebook.com",
		"twitter.com",
		"analytics.google.com",
		"doubleclick.net",
		"amazon.com",
		"cdn.",
		"cdnjs.",
	}
	
	for _, domain := range thirdPartyDomains {
		if strings.Contains(url, domain) {
			return true
		}
	}
	
	return false
}

func (a *Analyzer) calculateEstimatedPageLoadTime() float64 {
	if len(a.har.Log.Entries) == 0 {
		return 0
	}

	// Find the latest end time of all requests
	var maxEndTime time.Time
	var minStartTime time.Time = a.har.Log.Entries[0].StartedDateTime

	for _, entry := range a.har.Log.Entries {
		if entry.StartedDateTime.Before(minStartTime) {
			minStartTime = entry.StartedDateTime
		}
		
		endTime := entry.StartedDateTime.Add(time.Duration(entry.Time) * time.Millisecond)
		if endTime.After(maxEndTime) {
			maxEndTime = endTime
		}
	}

	return maxEndTime.Sub(minStartTime).Seconds() * 1000 // Convert to milliseconds
}

func (a *Analyzer) GenerateTimeline() []TimelineEvent {
	var events []TimelineEvent
	
	for i, entry := range a.har.Log.Entries {
		events = append(events, TimelineEvent{
			Index:       i,
			URL:         entry.Request.URL,
			Method:      entry.Request.Method,
			Status:      entry.Response.Status,
			StartTime:   entry.StartedDateTime,
			Duration:    entry.Time,
			Size:        entry.Response.Content.Size,
			ContentType: entry.Response.Content.MimeType,
		})
	}
	
	// Sort by start time
	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime.Before(events[j].StartTime)
	})
	
	return events
}

type TimelineEvent struct {
	Index       int
	URL         string
	Method      string
	Status      int
	StartTime   time.Time
	Duration    float64
	Size        int
	ContentType string
}