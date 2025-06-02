package tui

import (
	"fmt"
	"github.com/jlgore/hartea/internal/har"
	"github.com/jlgore/hartea/internal/report"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewMode int

const (
	TableView ViewMode = iota
	DetailView
	MetricsView
	TimelineView
	ComparisonView
	HelpView
)

type Model struct {
	harFiles      []*har.HAR
	analyzers     []*har.Analyzer
	currentFile   int
	currentView   ViewMode
	selectedEntry int
	
	// Components
	table       table.Model
	filter      textinput.Model
	
	// State
	width       int
	height      int
	loading     bool
	err         error
	showFilter  bool
	
	// Data
	entries     []har.Entry
	timeline    []har.TimelineEvent
	metrics     *har.Metrics
	comparison  *har.Comparison
	
	// Keybindings
	keys KeyMap
}

// Make render methods available
func (m Model) RenderTableView() string {
	var header string
	
	if len(m.harFiles) > 1 {
		header = titleStyle.Render(fmt.Sprintf("Hartea Analysis - Treasure Map %d/%d", m.currentFile+1, len(m.harFiles)))
	} else {
		header = titleStyle.Render("Hartea - Charting Digital Seas")
	}
	
	if m.metrics != nil {
		summary := fmt.Sprintf(
			"Requests: %d | Total Time: %.1fms | Total Size: %s | Errors: %d",
			m.metrics.TotalRequests,
			m.metrics.TotalTime,
			formatSize(int(m.metrics.TotalSize)),
			m.metrics.ErrorRequests,
		)
		header += "\n" + statusStyle.Render(summary)
	}
	
	var footer string
	if len(m.harFiles) > 1 {
		footer = "\n" + statusStyle.Render("Press ? for help, / to filter, m for metrics, t for timeline, c for comparison, e to export, q to quit")
	} else {
		footer = "\n" + statusStyle.Render("Press ? for help, / to filter, m for metrics, t for timeline, e to export, q to quit")
	}
	
	return header + "\n\n" + m.table.View() + footer
}

func (m Model) RenderFilter() string {
	header := titleStyle.Render("Filter Requests")
	prompt := "\n\n" + m.filter.View()
	help := "\n\nPress Enter to apply filter, Esc to cancel"
	
	return header + prompt + help
}

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	Back     key.Binding
	Filter   key.Binding
	Metrics    key.Binding
	Timeline   key.Binding
	Comparison key.Binding
	Export     key.Binding
	Help       key.Binding
	Quit       key.Binding
	Tab        key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "previous file"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next file"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Metrics: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "metrics"),
		),
		Timeline: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "timeline"),
		),
		Comparison: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "comparison"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export report"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch file"),
		),
	}
}

func NewModel(harFiles []*har.HAR) Model {
	analyzers := make([]*har.Analyzer, len(harFiles))
	for i, harFile := range harFiles {
		analyzers[i] = har.NewAnalyzer(harFile)
	}

	var entries []har.Entry
	var metrics *har.Metrics
	var timeline []har.TimelineEvent
	var comparison *har.Comparison
	
	if len(harFiles) > 0 {
		entries = harFiles[0].Log.Entries
		metrics = analyzers[0].CalculateMetrics()
		timeline = analyzers[0].GenerateTimeline()
	}
	
	// Create comparison if multiple files
	if len(harFiles) > 1 {
		allMetrics := make([]*har.Metrics, len(analyzers))
		fileNames := make([]string, len(harFiles))
		for i, analyzer := range analyzers {
			allMetrics[i] = analyzer.CalculateMetrics()
			fileNames[i] = fmt.Sprintf("File %d", i+1)
		}
		comparator := har.NewComparator(fileNames, allMetrics)
		comparison = comparator.Compare()
	}

	// Initialize table
	columns := []table.Column{
		{Title: "Method", Width: 8},
		{Title: "Status", Width: 6},
		{Title: "URL", Width: 60},
		{Title: "Time (ms)", Width: 10},
		{Title: "Size", Width: 10},
		{Title: "Type", Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	// Initialize filter
	filter := textinput.New()
	filter.Placeholder = "Filter requests..."
	filter.CharLimit = 256

	m := Model{
		harFiles:    harFiles,
		analyzers:   analyzers,
		currentFile: 0,
		currentView: TableView,
		table:       t,
		filter:      filter,
		entries:     entries,
		metrics:     metrics,
		timeline:    timeline,
		comparison:  comparison,
		keys:        DefaultKeyMap(),
	}

	m.updateTableRows()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(msg.Height - 10)
		
		// Update table column widths
		columns := m.table.Columns()
		if len(columns) > 0 {
			urlWidth := msg.Width - 60 // Reserve space for other columns
			if urlWidth > 30 {
				columns[2].Width = urlWidth
				m.table.SetColumns(columns)
			}
		}

	case tea.KeyMsg:
		if m.showFilter {
			switch {
			case key.Matches(msg, m.keys.Enter):
				m.showFilter = false
				m.filterEntries(m.filter.Value())
				return m, nil
			case key.Matches(msg, m.keys.Back):
				m.showFilter = false
				m.filter.SetValue("")
				return m, nil
			default:
				m.filter, cmd = m.filter.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Filter):
			m.showFilter = true
			m.filter.Focus()
			return m, nil

		case key.Matches(msg, m.keys.Tab):
			if len(m.harFiles) > 1 {
				m.currentFile = (m.currentFile + 1) % len(m.harFiles)
				m.switchFile()
			}
			return m, nil

		case key.Matches(msg, m.keys.Metrics):
			if m.currentView == MetricsView {
				m.currentView = TableView
			} else {
				m.currentView = MetricsView
			}
			return m, nil

		case key.Matches(msg, m.keys.Timeline):
			if m.currentView == TimelineView {
				m.currentView = TableView
			} else {
				m.currentView = TimelineView
			}
			return m, nil

		case key.Matches(msg, m.keys.Comparison):
			if len(m.harFiles) > 1 {
				if m.currentView == ComparisonView {
					m.currentView = TableView
				} else {
					m.currentView = ComparisonView
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Export):
			// Export reports
			go m.exportReports()
			return m, nil

		case key.Matches(msg, m.keys.Help):
			if m.currentView == HelpView {
				m.currentView = TableView
			} else {
				m.currentView = HelpView
			}
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			if m.currentView == TableView {
				m.selectedEntry = m.table.Cursor()
				m.currentView = DetailView
			}
			return m, nil

		case key.Matches(msg, m.keys.Back):
			if m.currentView != TableView {
				m.currentView = TableView
			}
			return m, nil
		}
	}

	if m.currentView == TableView && !m.showFilter {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.showFilter {
		return m.RenderFilter()
	}

	switch m.currentView {
	case DetailView:
		return m.renderDetailView()
	case MetricsView:
		return m.renderMetricsView()
	case TimelineView:
		return m.renderTimelineView()
	case ComparisonView:
		return m.renderComparisonView()
	case HelpView:
		return m.renderHelpView()
	default:
		return m.RenderTableView()
	}
}

func (m Model) renderDetailView() string {
	if m.selectedEntry >= len(m.entries) {
		return "No entry selected"
	}
	
	entry := m.entries[m.selectedEntry]
	
	var details []string
	
	// Header
	details = append(details, titleStyle.Render("Request Details"))
	details = append(details, "")
	
	// Request info
	details = append(details, headerStyle.Render("Request"))
	details = append(details, fmt.Sprintf("Method: %s", entry.Request.Method))
	details = append(details, fmt.Sprintf("URL: %s", entry.Request.URL))
	details = append(details, fmt.Sprintf("HTTP Version: %s", entry.Request.HTTPVersion))
	details = append(details, "")
	
	// Response info
	details = append(details, headerStyle.Render("Response"))
	details = append(details, fmt.Sprintf("Status: %d %s", entry.Response.Status, entry.Response.StatusText))
	details = append(details, fmt.Sprintf("Content Type: %s", entry.Response.Content.MimeType))
	details = append(details, fmt.Sprintf("Content Size: %s", formatSize(entry.Response.Content.Size)))
	if entry.Response.Content.Compression > 0 {
		details = append(details, fmt.Sprintf("Compression: %s saved", formatSize(entry.Response.Content.Compression)))
	}
	details = append(details, "")
	
	// Timing breakdown
	details = append(details, headerStyle.Render("Timing Breakdown"))
	details = append(details, fmt.Sprintf("Total Time: %.1fms", entry.Time))
	if entry.Timings.Blocked > 0 {
		details = append(details, fmt.Sprintf("Blocked: %dms", entry.Timings.Blocked))
	}
	if entry.Timings.DNS > 0 {
		details = append(details, fmt.Sprintf("DNS Lookup: %dms", entry.Timings.DNS))
	}
	if entry.Timings.Connect > 0 {
		details = append(details, fmt.Sprintf("TCP Connect: %dms", entry.Timings.Connect))
	}
	if entry.Timings.SSL > 0 {
		details = append(details, fmt.Sprintf("SSL Handshake: %dms", entry.Timings.SSL))
	}
	details = append(details, fmt.Sprintf("Send: %dms", entry.Timings.Send))
	details = append(details, fmt.Sprintf("Wait (TTFB): %dms", entry.Timings.Wait))
	details = append(details, fmt.Sprintf("Receive: %dms", entry.Timings.Receive))
	details = append(details, "")
	
	// Request headers (top 5)
	if len(entry.Request.Headers) > 0 {
		details = append(details, headerStyle.Render("Request Headers (Top 5)"))
		count := 0
		for _, header := range entry.Request.Headers {
			if count >= 5 {
				break
			}
			details = append(details, fmt.Sprintf("%s: %s", header.Name, truncateValue(header.Value, 60)))
			count++
		}
		if len(entry.Request.Headers) > 5 {
			details = append(details, fmt.Sprintf("... and %d more headers", len(entry.Request.Headers)-5))
		}
		details = append(details, "")
	}
	
	// Response headers (top 5)
	if len(entry.Response.Headers) > 0 {
		details = append(details, headerStyle.Render("Response Headers (Top 5)"))
		count := 0
		for _, header := range entry.Response.Headers {
			if count >= 5 {
				break
			}
			details = append(details, fmt.Sprintf("%s: %s", header.Name, truncateValue(header.Value, 60)))
			count++
		}
		if len(entry.Response.Headers) > 5 {
			details = append(details, fmt.Sprintf("... and %d more headers", len(entry.Response.Headers)-5))
		}
		details = append(details, "")
	}
	
	// Footer
	details = append(details, statusStyle.Render("Press Esc to go back"))
	
	return fmt.Sprintf("%s", details[0]) + "\n" + fmt.Sprintf("%s", details[1:])
}

func (m Model) renderMetricsView() string {
	if m.metrics == nil {
		return "No metrics available"
	}
	
	var content []string
	
	// Header
	content = append(content, titleStyle.Render("Performance Metrics"))
	content = append(content, "")
	
	// Core Web Vitals section
	content = append(content, headerStyle.Render("Core Performance Metrics"))
	ttfbStatus := ""
	if m.metrics.TTFB > 800 {
		ttfbStatus = " ⚠️  (Poor)"
	} else if m.metrics.TTFB > 200 {
		ttfbStatus = " ⚡ (Needs Improvement)"
	} else {
		ttfbStatus = " ✅ (Good)"
	}
	content = append(content, fmt.Sprintf("Time to First Byte (TTFB): %.1fms%s", m.metrics.TTFB, ttfbStatus))
	
	loadStatus := ""
	if m.metrics.PageLoadTime > 3000 {
		loadStatus = " ⚠️  (Poor)"
	} else if m.metrics.PageLoadTime > 1500 {
		loadStatus = " ⚡ (Needs Improvement)"
	} else {
		loadStatus = " ✅ (Good)"
	}
	content = append(content, fmt.Sprintf("Page Load Time: %.1fms%s", m.metrics.PageLoadTime, loadStatus))
	content = append(content, "")
	
	// Network metrics
	content = append(content, headerStyle.Render("Network Performance"))
	content = append(content, fmt.Sprintf("Average DNS Time: %.1fms", m.metrics.DNSTime))
	content = append(content, fmt.Sprintf("Average Connect Time: %.1fms", m.metrics.ConnectTime))
	if m.metrics.SSLTime > 0 {
		content = append(content, fmt.Sprintf("Average SSL Time: %.1fms", m.metrics.SSLTime))
	}
	content = append(content, "")
	
	// Request statistics
	content = append(content, headerStyle.Render("Request Statistics"))
	content = append(content, fmt.Sprintf("Total Requests: %d", m.metrics.TotalRequests))
	errorInfo := fmt.Sprintf("Error Requests: %d", m.metrics.ErrorRequests)
	if m.metrics.ErrorRequests > 0 {
		errorRate := float64(m.metrics.ErrorRequests) / float64(m.metrics.TotalRequests) * 100
		errorInfo += fmt.Sprintf(" (%.1f%%)", errorRate)
		if errorRate > 5 {
			errorInfo += " ⚠️"
		}
	}
	content = append(content, errorInfo)
	
	thirdPartyInfo := fmt.Sprintf("Third-party Requests: %d", m.metrics.ThirdPartyRequests)
	if m.metrics.TotalRequests > 0 {
		thirdPartyRate := float64(m.metrics.ThirdPartyRequests) / float64(m.metrics.TotalRequests) * 100
		thirdPartyInfo += fmt.Sprintf(" (%.1f%%)", thirdPartyRate)
	}
	content = append(content, thirdPartyInfo)
	content = append(content, "")
	
	// Cache efficiency
	content = append(content, headerStyle.Render("Cache Performance"))
	cacheInfo := fmt.Sprintf("Cache Hit Ratio: %.1f%%", m.metrics.CacheHitRatio)
	if m.metrics.CacheHitRatio < 30 {
		cacheInfo += " ⚠️  (Poor)"
	} else if m.metrics.CacheHitRatio < 60 {
		cacheInfo += " ⚡ (Needs Improvement)"
	} else {
		cacheInfo += " ✅ (Good)"
	}
	content = append(content, cacheInfo)
	content = append(content, "")
	
	// Size analysis
	content = append(content, headerStyle.Render("Size Analysis"))
	content = append(content, fmt.Sprintf("Total Transfer Size: %s", formatSize(int(m.metrics.TotalSize))))
	if m.metrics.TotalRequests > 0 {
		avgSize := m.metrics.TotalSize / int64(m.metrics.TotalRequests)
		content = append(content, fmt.Sprintf("Average Request Size: %s", formatSize(int(avgSize))))
	}
	content = append(content, "")
	
	// Performance recommendations
	content = append(content, headerStyle.Render("Recommendations"))
	
	if m.metrics.TTFB > 800 {
		content = append(content, "• Optimize server response time (TTFB > 800ms)")
	}
	if m.metrics.ErrorRequests > 0 {
		content = append(content, "• Fix HTTP errors to improve reliability")
	}
	if m.metrics.CacheHitRatio < 50 {
		content = append(content, "• Improve caching strategy for better performance")
	}
	if m.metrics.ThirdPartyRequests > m.metrics.TotalRequests/2 {
		content = append(content, "• Consider reducing third-party dependencies")
	}
	if m.metrics.TotalSize > 1024*1024*5 { // 5MB
		content = append(content, "• Optimize resource sizes and compression")
	}
	
	content = append(content, "")
	content = append(content, statusStyle.Render("Press Esc to go back"))
	
	return fmt.Sprintf("%s", content[0]) + "\n" + fmt.Sprintf("%s", content[1:])
}

func (m Model) renderHelpView() string {
	var help []string
	
	help = append(help, titleStyle.Render("Hartea - Navigator's Guide"))
	help = append(help, "")
	
	help = append(help, headerStyle.Render("Navigation"))
	help = append(help, "↑/k, ↓/j     Navigate up/down in table")
	help = append(help, "Enter        View request details")
	help = append(help, "Esc          Go back/cancel")
	help = append(help, "Tab          Switch between HAR files (if multiple)")
	help = append(help, "")
	
	help = append(help, headerStyle.Render("Views"))
	help = append(help, "m            Toggle metrics view")
	help = append(help, "t            Toggle timeline view")
	if len(m.harFiles) > 1 {
		help = append(help, "c            Toggle comparison view")
	}
	help = append(help, "e            Export reports (JSON/CSV/HTML/PDF)")
	help = append(help, "?            Toggle this help")
	help = append(help, "/            Filter requests")
	help = append(help, "")
	
	help = append(help, headerStyle.Render("Filtering"))
	help = append(help, "Type to filter by URL, method, or content type")
	help = append(help, "Examples: 'GET', 'javascript', 'api/', '404'")
	help = append(help, "")
	
	help = append(help, statusStyle.Render("Press q to quit, Esc to go back"))
	
	return fmt.Sprintf("%s", help[0]) + "\n" + fmt.Sprintf("%s", help[1:])
}

func (m Model) renderTimelineView() string {
	if len(m.entries) == 0 {
		return "No entries to display in timeline"
	}

	renderer := NewTimelineRenderer(m.width-4, m.height-10)
	return renderer.RenderWaterfall(m.entries, m.timeline)
}

type TimelineRenderer struct {
	width      int
	height     int
	pixelScale float64
	startTime  time.Time
	endTime    time.Time
}

func NewTimelineRenderer(width, height int) *TimelineRenderer {
	return &TimelineRenderer{
		width:  width,
		height: height,
	}
}

func (tr *TimelineRenderer) RenderWaterfall(entries []har.Entry, timeline []har.TimelineEvent) string {
	if len(timeline) == 0 {
		return "No timeline data available"
	}

	// Calculate time bounds
	tr.startTime = timeline[0].StartTime
	tr.endTime = timeline[0].StartTime
	
	for _, event := range timeline {
		if event.StartTime.Before(tr.startTime) {
			tr.startTime = event.StartTime
		}
		endTime := event.StartTime.Add(time.Duration(event.Duration) * time.Millisecond)
		if endTime.After(tr.endTime) {
			tr.endTime = endTime
		}
	}

	totalDuration := tr.endTime.Sub(tr.startTime).Seconds() * 1000
	if totalDuration <= 0 {
		totalDuration = 1000
	}

	chartWidth := tr.width - 35
	if chartWidth < 20 {
		chartWidth = 20
	}
	
	tr.pixelScale = totalDuration / float64(chartWidth)

	var output []string
	
	output = append(output, titleStyle.Render("Request Timeline (Waterfall Chart)"))
	output = append(output, "")
	
	output = append(output, tr.renderTimeScale(chartWidth))
	output = append(output, "")
	
	maxEntries := tr.height - 8
	entriesToShow := len(timeline)
	if entriesToShow > maxEntries {
		entriesToShow = maxEntries
	}
	
	for i := 0; i < entriesToShow; i++ {
		event := timeline[i]
		output = append(output, tr.renderRequestBar(event, chartWidth, i))
	}
	
	if len(timeline) > maxEntries {
		output = append(output, fmt.Sprintf("... and %d more requests", len(timeline)-maxEntries))
	}
	
	output = append(output, "")
	output = append(output, tr.renderLegend())
	output = append(output, "")
	output = append(output, statusStyle.Render("Press Esc to go back"))
	
	return strings.Join(output, "\n")
}

func (tr *TimelineRenderer) renderTimeScale(chartWidth int) string {
	scale := strings.Repeat(" ", 30)
	
	scaleLine := make([]rune, chartWidth)
	for i := range scaleLine {
		scaleLine[i] = '─'
	}
	
	totalMs := tr.endTime.Sub(tr.startTime).Seconds() * 1000
	markers := []float64{0, 0.25, 0.5, 0.75, 1.0}
	
	for _, marker := range markers {
		pos := int(float64(chartWidth) * marker)
		if pos < chartWidth {
			scaleLine[pos] = '┬'
		}
	}
	
	scale += string(scaleLine)
	scale += "\n" + strings.Repeat(" ", 30)
	
	labelLine := make([]rune, chartWidth)
	for i := range labelLine {
		labelLine[i] = ' '
	}
	
	for _, marker := range markers {
		pos := int(float64(chartWidth) * marker)
		timeMs := totalMs * marker
		timeLabel := fmt.Sprintf("%.0fms", timeMs)
		
		labelStart := pos - len(timeLabel)/2
		if labelStart < 0 {
			labelStart = 0
		}
		if labelStart+len(timeLabel) >= chartWidth {
			labelStart = chartWidth - len(timeLabel)
		}
		
		if labelStart >= 0 {
			for j, char := range timeLabel {
				if labelStart+j < chartWidth {
					labelLine[labelStart+j] = char
				}
			}
		}
	}
	
	scale += string(labelLine)
	return scale
}

func (tr *TimelineRenderer) renderRequestBar(event har.TimelineEvent, chartWidth, index int) string {
	label := tr.formatRequestLabel(event)
	if len(label) > 28 {
		label = label[:25] + "..."
	}
	
	bar := fmt.Sprintf("%-30s", label)
	
	requestStart := event.StartTime.Sub(tr.startTime).Seconds() * 1000
	requestDuration := event.Duration
	
	startPos := int(requestStart / tr.pixelScale)
	duration := int(requestDuration / tr.pixelScale)
	
	if duration < 1 {
		duration = 1
	}
	
	if startPos >= chartWidth {
		startPos = chartWidth - 1
	}
	if startPos+duration > chartWidth {
		duration = chartWidth - startPos
	}
	
	timeline := make([]rune, chartWidth)
	for i := range timeline {
		timeline[i] = ' '
	}
	
	barChar, barStyle := tr.getBarStyle(event)
	for i := startPos; i < startPos+duration && i < chartWidth; i++ {
		timeline[i] = barChar
	}
	
	if startPos+duration < chartWidth {
		if event.Status >= 400 {
			timeline[startPos+duration] = '✗'
		} else if event.Status >= 300 {
			timeline[startPos+duration] = '↻'
		} else {
			timeline[startPos+duration] = '✓'
		}
	}
	
	timelineStr := string(timeline)
	timelineStr = barStyle.Render(timelineStr)
	
	bar += timelineStr
	bar += fmt.Sprintf(" %s %.1fms", tr.getStatusIcon(event.Status), event.Duration)
	
	return bar
}

func (tr *TimelineRenderer) formatRequestLabel(event har.TimelineEvent) string {
	parts := strings.Split(event.URL, "/")
	filename := parts[len(parts)-1]
	if filename == "" || filename == event.URL {
		if strings.Contains(event.URL, "://") {
			urlParts := strings.Split(event.URL, "://")
			if len(urlParts) > 1 {
				domainParts := strings.Split(urlParts[1], "/")
				filename = domainParts[0]
			}
		}
	}
	
	if strings.Contains(filename, "?") {
		filename = strings.Split(filename, "?")[0]
	}
	
	return fmt.Sprintf("%s %s", event.Method, filename)
}

func (tr *TimelineRenderer) getBarStyle(event har.TimelineEvent) (rune, lipgloss.Style) {
	if event.Status >= 400 {
		return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	}
	
	if event.Status >= 300 {
		return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	}
	
	if strings.Contains(event.ContentType, "html") {
		return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	} else if strings.Contains(event.ContentType, "javascript") {
		return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	} else if strings.Contains(event.ContentType, "css") {
		return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	} else if strings.Contains(event.ContentType, "image") {
		return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	} else if strings.Contains(event.ContentType, "json") {
		return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	} else if strings.Contains(event.ContentType, "font") {
		return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	}
	
	return '█', lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
}

func (tr *TimelineRenderer) getStatusIcon(status int) string {
	if status >= 400 {
		return "❌"
	} else if status >= 300 {
		return "🔄"
	} else if status >= 200 {
		return "✅"
	}
	return "❓"
}

func (tr *TimelineRenderer) renderLegend() string {
	var legend []string
	
	legend = append(legend, headerStyle.Render("Legend:"))
	
	htmlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	jsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	cssStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	imgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	apiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	fontStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	
	legend = append(legend, fmt.Sprintf("%s HTML  %s JS  %s CSS  %s Images  %s API/JSON  %s Fonts",
		htmlStyle.Render("█"),
		jsStyle.Render("█"),
		cssStyle.Render("█"),
		imgStyle.Render("█"),
		apiStyle.Render("█"),
		fontStyle.Render("█")))
	
	legend = append(legend, "Status: ✅ Success  🔄 Redirect  ❌ Error")
	
	return strings.Join(legend, "\n")
}

func (m Model) renderComparisonView() string {
	if m.comparison == nil {
		return "No comparison data available. Load multiple HAR files to compare."
	}

	var content []string
	
	// Header
	content = append(content, titleStyle.Render(fmt.Sprintf("Performance Comparison (%d files)", len(m.harFiles))))
	content = append(content, "")
	
	// Summary
	summary := m.comparison.Summary
	summaryText := fmt.Sprintf("📊 %d Better | %d Worse | %d Unchanged (of %d metrics)", 
		summary.BetterCount, summary.WorseCount, summary.UnchangedCount, summary.TotalMetrics)
	content = append(content, headerStyle.Render(summaryText))
	content = append(content, "")
	
	// Metrics table header
	header := fmt.Sprintf("%-25s", "Metric")
	for i, file := range m.comparison.Files {
		if i == 0 {
			header += fmt.Sprintf("%-15s", file+" (Base)")
		} else {
			header += fmt.Sprintf("%-20s", file)
		}
	}
	content = append(content, headerStyle.Render(header))
	content = append(content, strings.Repeat("─", len(header)))
	
	// Metrics comparison
	for _, diff := range m.comparison.Differences {
		row := fmt.Sprintf("%-25s", diff.Name)
		
		for i, value := range diff.Values {
			valueStr := fmt.Sprintf("%v", value)
			if i == 0 {
				row += fmt.Sprintf("%-15s", valueStr)
			} else {
				change := diff.Changes[i]
				improvement := diff.Improvements[i]
				
				// Add styling based on improvement
				changeStyled := change
				if change != "Baseline" && change != "No change" {
					if improvement {
						changeStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(change + " ✅")
					} else {
						changeStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(change + " ⚠️")
					}
				}
				
				combined := fmt.Sprintf("%s (%s)", valueStr, changeStyled)
				row += fmt.Sprintf("%-20s", combined)
			}
		}
		
		content = append(content, row)
	}
	
	content = append(content, "")
	content = append(content, "")
	
	// Insights
	content = append(content, headerStyle.Render("Key Insights"))
	insights := m.generateInsights()
	for _, insight := range insights {
		content = append(content, "• "+insight)
	}
	
	content = append(content, "")
	content = append(content, statusStyle.Render("Press Esc to go back"))
	
	return strings.Join(content, "\n")
}

func (m Model) generateInsights() []string {
	if m.comparison == nil || len(m.comparison.Differences) == 0 {
		return []string{"No insights available"}
	}

	var insights []string
	
	// Analyze load time changes
	for _, diff := range m.comparison.Differences {
		if diff.Name == "Total Load Time" && len(diff.Changes) > 1 {
			change := diff.Changes[1]
			if strings.Contains(change, "-") && diff.Improvements[1] {
				insights = append(insights, "Page load time improved significantly")
			} else if strings.Contains(change, "+") && !diff.Improvements[1] {
				insights = append(insights, "Page load time regressed - investigate performance")
			}
		}
		
		if diff.Name == "Error Requests" && len(diff.Changes) > 1 {
			change := diff.Changes[1]
			if change == "No change" || strings.Contains(change, "-") {
				insights = append(insights, "Error rate remained stable or improved")
			} else if strings.Contains(change, "+") {
				insights = append(insights, "Error rate increased - check for new issues")
			}
		}
		
		if diff.Name == "Cache Hit Ratio" && len(diff.Changes) > 1 {
			change := diff.Changes[1]
			if strings.Contains(change, "+") && diff.Improvements[1] {
				insights = append(insights, "Cache efficiency improved")
			} else if strings.Contains(change, "-") && !diff.Improvements[1] {
				insights = append(insights, "Cache efficiency decreased")
			}
		}
		
		if diff.Name == "Total Transfer Size" && len(diff.Changes) > 1 {
			change := diff.Changes[1]
			if strings.Contains(change, "-") && diff.Improvements[1] {
				insights = append(insights, "Transfer size optimized")
			} else if strings.Contains(change, "+") && !diff.Improvements[1] {
				insights = append(insights, "Transfer size increased - check for new assets")
			}
		}
	}
	
	if len(insights) == 0 {
		insights = append(insights, "Performance appears stable across files")
	}
	
	return insights
}

func (m Model) exportReports() {
	generator := report.NewGenerator(m.harFiles, m.analyzers, m.comparison)
	
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	baseFilename := fmt.Sprintf("har-analysis-%s", timestamp)
	
	// Export all formats
	formats := []struct {
		extension string
		exportFunc func(string) error
	}{
		{".json", func(filename string) error { return generator.ExportJSON(filename, false) }},
		{".csv", generator.ExportCSV},
		{".html", generator.ExportHTML},
		{".pdf", generator.ExportPDF},
	}
	
	for _, format := range formats {
		filename := baseFilename + format.extension
		if err := format.exportFunc(filename); err != nil {
			// In a real implementation, you might want to show this error in the UI
			continue
		}
	}
}

func truncateValue(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen-3] + "..."
}

func (m *Model) updateTableRows() {
	if len(m.entries) == 0 {
		return
	}

	rows := make([]table.Row, len(m.entries))
	for i, entry := range m.entries {
		size := formatSize(entry.Response.Content.Size)
		contentType := entry.Response.Content.MimeType
		if contentType == "" {
			contentType = "unknown"
		}
		if len(contentType) > 15 {
			contentType = contentType[:12] + "..."
		}

		rows[i] = table.Row{
			entry.Request.Method,
			fmt.Sprintf("%d", entry.Response.Status),
			truncateURL(entry.Request.URL, 60),
			fmt.Sprintf("%.1f", entry.Time),
			size,
			contentType,
		}
	}
	m.table.SetRows(rows)
}

func (m *Model) switchFile() {
	if m.currentFile < len(m.harFiles) {
		m.entries = m.harFiles[m.currentFile].Log.Entries
		m.metrics = m.analyzers[m.currentFile].CalculateMetrics()
		m.timeline = m.analyzers[m.currentFile].GenerateTimeline()
		m.updateTableRows()
		m.selectedEntry = 0
		m.table.GotoTop()
	}
}

func (m *Model) filterEntries(filterText string) {
	if filterText == "" {
		m.entries = m.harFiles[m.currentFile].Log.Entries
	} else {
		var filtered []har.Entry
		for _, entry := range m.harFiles[m.currentFile].Log.Entries {
			if matchesFilter(entry, filterText) {
				filtered = append(filtered, entry)
			}
		}
		m.entries = filtered
	}
	m.updateTableRows()
	m.table.GotoTop()
}

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))
		
	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("242"))
)

func formatSize(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	}
}

func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}

func matchesFilter(entry har.Entry, filter string) bool {
	// Simple case-insensitive matching
	filter = fmt.Sprintf("%s", filter)
	url := fmt.Sprintf("%s", entry.Request.URL)
	method := fmt.Sprintf("%s", entry.Request.Method)
	contentType := fmt.Sprintf("%s", entry.Response.Content.MimeType)
	
	return contains(url, filter) || 
		   contains(method, filter) || 
		   contains(contentType, filter)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(substr) == 0 || 
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if toLower(a[i]) != toLower(b[i]) {
			return false
		}
	}
	return true
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}