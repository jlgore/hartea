package har

import (
	"fmt"
	"math"
)

type Comparison struct {
	Files        []string
	Metrics      []*Metrics
	Differences  []MetricDifference
	Summary      ComparisonSummary
}

type MetricDifference struct {
	Name        string
	Values      []interface{}
	Changes     []string
	Improvements []bool
}

type ComparisonSummary struct {
	BetterCount   int
	WorseCount    int
	UnchangedCount int
	TotalMetrics  int
}

type Comparator struct {
	files    []string
	metrics  []*Metrics
}

func NewComparator(files []string, metrics []*Metrics) *Comparator {
	return &Comparator{
		files:   files,
		metrics: metrics,
	}
}

func (c *Comparator) Compare() *Comparison {
	if len(c.metrics) < 2 {
		return &Comparison{
			Files:   c.files,
			Metrics: c.metrics,
		}
	}

	comparison := &Comparison{
		Files:   c.files,
		Metrics: c.metrics,
	}

	// Compare key metrics
	comparison.Differences = []MetricDifference{
		c.compareFloat("Total Load Time", "ms", extractPageLoadTime),
		c.compareFloat("Time to First Byte", "ms", extractTTFB),
		c.compareFloat("Average DNS Time", "ms", extractDNSTime),
		c.compareFloat("Average Connect Time", "ms", extractConnectTime),
		c.compareFloat("Average SSL Time", "ms", extractSSLTime),
		c.compareInt("Total Requests", "", extractTotalRequests),
		c.compareInt("Error Requests", "", extractErrorRequests),
		c.compareInt("Third-party Requests", "", extractThirdPartyRequests),
		c.compareFloat("Cache Hit Ratio", "%", extractCacheHitRatio),
		c.compareSize("Total Transfer Size", extractTotalSize),
	}

	// Calculate summary
	comparison.Summary = c.calculateSummary(comparison.Differences)

	return comparison
}

func (c *Comparator) compareFloat(name, unit string, extractor func(*Metrics) float64) MetricDifference {
	values := make([]interface{}, len(c.metrics))
	changes := make([]string, len(c.metrics))
	improvements := make([]bool, len(c.metrics))

	for i, metric := range c.metrics {
		value := extractor(metric)
		values[i] = fmt.Sprintf("%.1f%s", value, unit)
		
		if i > 0 {
			baseValue := extractor(c.metrics[0])
			change := value - baseValue
			changePercent := 0.0
			if baseValue != 0 {
				changePercent = (change / baseValue) * 100
			}
			
			if math.Abs(changePercent) < 0.1 {
				changes[i] = "No change"
				improvements[i] = false
			} else if changePercent > 0 {
				changes[i] = fmt.Sprintf("+%.1f%%", changePercent)
				improvements[i] = isImprovementFloat(name, change)
			} else {
				changes[i] = fmt.Sprintf("%.1f%%", changePercent)
				improvements[i] = isImprovementFloat(name, change)
			}
		} else {
			changes[i] = "Baseline"
			improvements[i] = false
		}
	}

	return MetricDifference{
		Name:         name,
		Values:       values,
		Changes:      changes,
		Improvements: improvements,
	}
}

func (c *Comparator) compareInt(name, unit string, extractor func(*Metrics) int) MetricDifference {
	values := make([]interface{}, len(c.metrics))
	changes := make([]string, len(c.metrics))
	improvements := make([]bool, len(c.metrics))

	for i, metric := range c.metrics {
		value := extractor(metric)
		if unit != "" {
			values[i] = fmt.Sprintf("%d %s", value, unit)
		} else {
			values[i] = fmt.Sprintf("%d", value)
		}
		
		if i > 0 {
			baseValue := extractor(c.metrics[0])
			change := value - baseValue
			changePercent := 0.0
			if baseValue != 0 {
				changePercent = (float64(change) / float64(baseValue)) * 100
			}
			
			if change == 0 {
				changes[i] = "No change"
				improvements[i] = false
			} else if change > 0 {
				changes[i] = fmt.Sprintf("+%d (+%.1f%%)", change, changePercent)
				improvements[i] = isImprovementInt(name, change)
			} else {
				changes[i] = fmt.Sprintf("%d (%.1f%%)", change, changePercent)
				improvements[i] = isImprovementInt(name, change)
			}
		} else {
			changes[i] = "Baseline"
			improvements[i] = false
		}
	}

	return MetricDifference{
		Name:         name,
		Values:       values,
		Changes:      changes,
		Improvements: improvements,
	}
}

func (c *Comparator) compareSize(name string, extractor func(*Metrics) int64) MetricDifference {
	values := make([]interface{}, len(c.metrics))
	changes := make([]string, len(c.metrics))
	improvements := make([]bool, len(c.metrics))

	for i, metric := range c.metrics {
		value := extractor(metric)
		values[i] = formatSize(int(value))
		
		if i > 0 {
			baseValue := extractor(c.metrics[0])
			change := value - baseValue
			changePercent := 0.0
			if baseValue != 0 {
				changePercent = (float64(change) / float64(baseValue)) * 100
			}
			
			if change == 0 {
				changes[i] = "No change"
				improvements[i] = false
			} else if change > 0 {
				changes[i] = fmt.Sprintf("+%s (+%.1f%%)", formatSize(int(change)), changePercent)
				improvements[i] = change < 0 // Smaller size is better
			} else {
				changes[i] = fmt.Sprintf("-%s (%.1f%%)", formatSize(int(-change)), changePercent)
				improvements[i] = change < 0 // Smaller size is better
			}
		} else {
			changes[i] = "Baseline"
			improvements[i] = false
		}
	}

	return MetricDifference{
		Name:         name,
		Values:       values,
		Changes:      changes,
		Improvements: improvements,
	}
}

func (c *Comparator) calculateSummary(differences []MetricDifference) ComparisonSummary {
	var better, worse, unchanged int
	
	for _, diff := range differences {
		for i := 1; i < len(diff.Improvements); i++ {
			if diff.Changes[i] == "No change" {
				unchanged++
			} else if diff.Improvements[i] {
				better++
			} else {
				worse++
			}
		}
	}

	return ComparisonSummary{
		BetterCount:    better,
		WorseCount:     worse,
		UnchangedCount: unchanged,
		TotalMetrics:   better + worse + unchanged,
	}
}

// Extractor functions
func extractPageLoadTime(m *Metrics) float64   { return m.PageLoadTime }
func extractTTFB(m *Metrics) float64          { return m.TTFB }
func extractDNSTime(m *Metrics) float64       { return m.DNSTime }
func extractConnectTime(m *Metrics) float64   { return m.ConnectTime }
func extractSSLTime(m *Metrics) float64       { return m.SSLTime }
func extractTotalRequests(m *Metrics) int     { return m.TotalRequests }
func extractErrorRequests(m *Metrics) int     { return m.ErrorRequests }
func extractThirdPartyRequests(m *Metrics) int { return m.ThirdPartyRequests }
func extractCacheHitRatio(m *Metrics) float64 { return m.CacheHitRatio }
func extractTotalSize(m *Metrics) int64       { return m.TotalSize }

// Improvement detection
func isImprovementFloat(metricName string, change float64) bool {
	// For timing metrics, smaller is better
	timingMetrics := []string{
		"Total Load Time", "Time to First Byte", "Average DNS Time",
		"Average Connect Time", "Average SSL Time",
	}
	
	for _, timing := range timingMetrics {
		if metricName == timing {
			return change < 0
		}
	}
	
	// For cache hit ratio, higher is better
	if metricName == "Cache Hit Ratio" {
		return change > 0
	}
	
	return false
}

func isImprovementInt(metricName string, change int) bool {
	// For error requests and third-party requests, fewer is better
	if metricName == "Error Requests" || metricName == "Third-party Requests" {
		return change < 0
	}
	
	// For total requests, depends on context - we'll consider it neutral
	return false
}

func formatSize(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	}
}