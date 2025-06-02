# HAR Analyzer

A terminal-based HAR (HTTP Archive) file analyzer built with Go and Charm's Bubbletea framework. Analyze web performance, compare multiple HAR files, and generate detailed reports directly in your terminal.

## Features

### 🔍 **Comprehensive Analysis**
- **Performance Metrics**: TTFB, Page Load Time, Core Web Vitals
- **Network Analysis**: DNS lookup, TCP connection, SSL handshake timings
- **Request Statistics**: Total requests, error rates, third-party analysis
- **Cache Efficiency**: Hit ratios and optimization recommendations
- **Resource Breakdown**: Analysis by content type (JS, CSS, images, etc.)

### 📊 **Interactive Interface**
- **Table View**: Sortable and filterable list of all HTTP requests
- **Detail View**: In-depth request/response analysis with timing breakdown
- **Metrics Dashboard**: Performance overview with recommendations
- **Timeline View**: ASCII waterfall chart like Chrome DevTools
- **Comparison View**: Side-by-side performance analysis of multiple HAR files
- **Report Export**: Generate professional reports in JSON, CSV, HTML, and PDF formats
- **Multi-file Support**: Load and compare multiple HAR files seamlessly

### ⚡ **Performance Optimized**
- **Streaming JSON Parser**: Handles large HAR files efficiently
- **Buffered I/O**: Optimized for files of any size
- **Memory Efficient**: Pagination and lazy loading for large datasets

## Installation

### 📦 **Pre-built Binaries**
Download the latest release for your platform from [GitHub Releases](https://github.com/YOUR_USERNAME/har-analyzer/releases).

### 🐳 **Docker**
```bash
# Pull and run from GitHub Container Registry
docker pull ghcr.io/YOUR_USERNAME/har-analyzer:latest

# Run with HAR files from current directory
docker run --rm -v $(pwd):/data ghcr.io/YOUR_USERNAME/har-analyzer:latest /data/example.har

# Compare multiple files
docker run --rm -v $(pwd):/data ghcr.io/YOUR_USERNAME/har-analyzer:latest /data/before.har /data/after.har
```

### 🍺 **Homebrew (macOS/Linux)**
```bash
# Add the tap
brew tap YOUR_USERNAME/tap

# Install
brew install har-analyzer
```

### 📦 **Package Managers**
```bash
# Debian/Ubuntu
wget https://github.com/YOUR_USERNAME/har-analyzer/releases/download/v1.0.0/har-analyzer_1.0.0_linux_amd64.deb
sudo dpkg -i har-analyzer_1.0.0_linux_amd64.deb

# Red Hat/CentOS/Fedora
wget https://github.com/YOUR_USERNAME/har-analyzer/releases/download/v1.0.0/har-analyzer_1.0.0_linux_amd64.rpm
sudo rpm -i har-analyzer_1.0.0_linux_amd64.rpm

# Alpine Linux
wget https://github.com/YOUR_USERNAME/har-analyzer/releases/download/v1.0.0/har-analyzer_1.0.0_linux_amd64.apk
sudo apk add har-analyzer_1.0.0_linux_amd64.apk
```

### 🛠️ **Build from Source**
```bash
# Clone and build
git clone https://github.com/YOUR_USERNAME/har-analyzer
cd har-analyzer
go build -o har-analyzer ./cmd/main.go
```

## Usage

### Basic Analysis
```bash
# Analyze a single HAR file
./har-analyzer example.har

# Analyze multiple HAR files for comparison
./har-analyzer before.har after.har
```

### Navigation
- **↑/k, ↓/j**: Navigate up/down in table
- **Enter**: View request details
- **Esc**: Go back/cancel
- **Tab**: Switch between HAR files (if multiple)
- **m**: Toggle metrics view
- **t**: Toggle timeline view
- **c**: Toggle comparison view (when multiple files loaded)
- **e**: Export reports (JSON/CSV/HTML/PDF)
- **?**: Toggle help
- **/**: Filter requests
- **q**: Quit

### Filtering
Filter requests by typing after pressing `/`:
- `GET` - Show only GET requests
- `javascript` - Show only JavaScript files
- `api/` - Show only API calls
- `404` - Show only 404 errors

### Comparison Analysis
Compare multiple HAR files to analyze performance changes:

```bash
# Before vs After deployment comparison
./har-analyzer before-deploy.har after-deploy.har

# A/B testing analysis
./har-analyzer variant-a.har variant-b.har

# Cross-environment comparison
./har-analyzer staging.har production.har
```

Press **c** when multiple files are loaded to see:
- **Side-by-side metrics comparison** with percentage changes
- **Performance regression/improvement detection**
- **Automated insights** and recommendations
- **Color-coded indicators**: ✅ Improvements, ⚠️ Regressions
- **Summary statistics**: Better/Worse/Unchanged metrics count

Example comparison output:
```
Performance Comparison (2 files)
📊 6 Better | 1 Worse | 3 Unchanged (of 10 metrics)

Metric                 File 1 (Base)   File 2
Total Load Time        2500.0ms        1800.0ms (-28.0% ✅)
Time to First Byte     450.0ms         280.0ms (-37.8% ✅)
Total Requests         45              52 (+7 +15.6% ⚠️)
Cache Hit Ratio        45.0%           78.0% (+33.0% ✅)

Key Insights:
• Page load time improved significantly
• Cache efficiency improved
• Error rate remained stable or improved
```

### Report Export
Generate professional reports in multiple formats by pressing **e**:

**Supported Formats:**
- **JSON**: Machine-readable data for integration with other tools
- **CSV**: Spreadsheet-compatible metrics for data analysis
- **HTML**: Styled web report with interactive elements and visual indicators
- **PDF**: Professional document with charts, tables, and recommendations

**Report Contents:**
- **Executive Summary**: Key performance indicators and overall health
- **Detailed Metrics**: Complete breakdown of all timing and size metrics
- **Performance Comparison**: Side-by-side analysis (when multiple files loaded)
- **Recommendations**: Automated insights and optimization suggestions
- **Visual Indicators**: Color-coded status indicators for quick assessment

Example exported files:
```
har-analysis-2024-01-15_14-30-25.json  # Raw data export
har-analysis-2024-01-15_14-30-25.csv   # Metrics spreadsheet
har-analysis-2024-01-15_14-30-25.html  # Interactive web report
har-analysis-2024-01-15_14-30-25.pdf   # Professional document
```

**PDF Report Features:**
- **Professional Layout**: Clean, branded design suitable for stakeholders
- **Color-coded Metrics**: Visual performance indicators (green/yellow/red)
- **Comparison Tables**: Side-by-side analysis with improvement/regression markers
- **Automated Recommendations**: Actionable insights based on performance data
- **Summary Dashboard**: Executive overview with key metrics
- **Multi-page Support**: Comprehensive analysis without space constraints

## Architecture

```
har-analyzer/
├── cmd/
│   └── main.go                 # CLI entry point
├── internal/
│   ├── har/
│   │   ├── types.go           # HAR data structures
│   │   ├── parser.go          # HAR file parsing
│   │   └── analyzer.go        # Performance analysis
│   └── tui/
│       ├── model.go           # Main Bubbletea model
│       └── views/             # UI view components
└── go.mod
```

## Performance Metrics

### Core Web Vitals
- **Time to First Byte (TTFB)**
  - ✅ Good: ≤200ms
  - ⚡ Needs Improvement: 200-800ms  
  - ⚠️ Poor: >800ms

- **Page Load Time**
  - ✅ Good: ≤1.5s
  - ⚡ Needs Improvement: 1.5-3s
  - ⚠️ Poor: >3s

### Network Performance
- Average DNS lookup time
- TCP connection establishment
- SSL handshake duration
- Request/response timing breakdown

### Cache Analysis
- Cache hit ratio calculation
- Resource optimization opportunities
- Compression efficiency analysis

## Technical Implementation

### HAR Parsing
- **Streaming JSON Parser**: Uses `json.NewDecoder` for memory efficiency
- **Buffered I/O**: 64KB buffer for optimal file reading performance
- **Validation**: Comprehensive HAR format validation
- **Error Handling**: Graceful handling of malformed HAR files

### TUI Framework
- **Bubbletea**: Modern terminal UI framework following The Elm Architecture
- **Bubbles Components**: Pre-built UI components (table, textinput, etc.)
- **Lipgloss**: Styling and layout system for terminal interfaces
- **Responsive Design**: Adapts to different terminal sizes

### Performance Optimizations
- **Lazy Loading**: Load data on-demand for large HAR files
- **Pagination**: Display only visible subset of data
- **Efficient Filtering**: Fast string matching without regex overhead
- **Memory Management**: Careful handling of large datasets

## Example Output

```
HAR Analysis - File 1/2
Requests: 127 | Total Time: 2847.3ms | Total Size: 1.2MB | Errors: 0

┌────────┬────────┬──────────────────────────────────────────────┬──────────┬────────┬─────────────────┐
│ Method │ Status │ URL                                          │ Time     │ Size   │ Type            │
├────────┼────────┼──────────────────────────────────────────────┼──────────┼────────┼─────────────────┤
│ GET    │ 200    │ https://example.com/                         │ 234.5    │ 15.2KB │ html            │
│ GET    │ 200    │ https://example.com/assets/app.js            │ 145.2    │ 87.5KB │ javascript      │
│ GET    │ 200    │ https://example.com/assets/style.css         │ 89.1     │ 23.1KB │ css             │
└────────┴────────┴──────────────────────────────────────────────┴──────────┴────────┴─────────────────┘

Press ? for help, / to filter, m for metrics, q to quit
```

## Dependencies

- **Bubbletea**: Terminal user interface framework
- **Bubbles**: Pre-built UI components
- **Lipgloss**: Styling and layout
- **Standard Library**: JSON parsing, file I/O, string manipulation

## CI/CD Pipeline

The project includes a comprehensive CI/CD pipeline with:

### 🔄 **Continuous Integration**
- **Multi-Go Version Testing**: Tests against Go 1.22, 1.23, and 1.24
- **Cross-Platform Builds**: Linux, macOS, Windows (amd64, arm64)
- **Code Quality**: Static analysis with `staticcheck`, `golint`, and `go vet`
- **Security Scanning**: Vulnerability scans with Trivy and Gosec
- **Coverage Reports**: Automated test coverage with Codecov

### 🚀 **Automated Releases**
- **GoReleaser**: Automated cross-platform binary builds
- **GitHub Releases**: Automatic release creation with changelogs
- **Package Distribution**: Debian/RPM/Alpine packages
- **Homebrew Formula**: Automatic tap updates
- **Container Images**: Multi-arch Docker images (amd64/arm64)

### 🐳 **Container Registry**
- **GitHub Container Registry**: `ghcr.io/YOUR_USERNAME/har-analyzer`
- **Multi-Architecture**: Native amd64 and arm64 support
- **Security**: Regular vulnerability scans and updates
- **Size Optimized**: Minimal scratch-based images

### 📋 **Software Bill of Materials (SBOM)**
- **Dependency Tracking**: Complete SBOM generation for security compliance
- **Supply Chain Security**: Signed releases with Cosign
- **Vulnerability Monitoring**: Automated security updates via Dependabot

## Development

### 🧪 **Testing**
```bash
# Run tests
go test -v ./...

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out
```

### 🔧 **Local Development**
```bash
# Install development dependencies
go mod download

# Run static analysis
staticcheck ./...
golint ./...
go vet ./...

# Build with version info
go build -ldflags="-X main.version=dev" -o har-analyzer ./cmd/main.go
```

### 🐳 **Docker Development**
```bash
# Build locally
docker build -t har-analyzer:dev .

# Test the container
docker run --rm -v $(pwd):/data har-analyzer:dev /data/example.har
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run the test suite (`go test ./...`)
6. Commit your changes (`git commit -m 'feat: add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Submit a pull request

### 📝 **Commit Convention**
This project follows [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code style changes
- `refactor:` Code refactoring
- `test:` Test additions or changes
- `chore:` Maintenance tasks

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Security

For security issues, please email security@yourproject.com or use [GitHub Security Advisories](https://github.com/YOUR_USERNAME/har-analyzer/security/advisories).

## Acknowledgments

- Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI framework
- PDF generation powered by [gofpdf](https://github.com/jung-kurt/gofpdf)
- Inspired by Chrome DevTools Network panel

## Performance Tips

- For very large HAR files (>100MB), consider using `--buffer-size` flag to increase buffer size
- Use filtering to focus on specific request types or domains
- Multiple file comparison works best with files of similar size and structure
- Terminal width affects table column sizing - wider terminals show more data

## Future Enhancements

- [ ] JSON/CSV/HTML report export
- [ ] Timeline waterfall visualization
- [ ] Custom performance budgets
- [ ] Plugin system for custom analysis
- [ ] Real-time HAR file monitoring
- [ ] Automated performance regression detection