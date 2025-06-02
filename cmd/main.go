package main

import (
	"fmt"
	"github.com/jlgore/hartea/internal/har"
	"github.com/jlgore/hartea/internal/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Build information (set via ldflags)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Handle version flag
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("hartea %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		fmt.Printf("built by: %s\n", builtBy)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Println("Hartea " + version)
		fmt.Println("Advanced terminal-based HAR file analysis tool - Ahoy Matey!")
		fmt.Println("")
		fmt.Println("Usage: hartea <har-file1> [har-file2] ...")
		fmt.Println("       hartea --version")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  hartea example.har                    # Analyze single file")
		fmt.Println("  hartea before.har after.har          # Compare two files")
		fmt.Println("  hartea *.har                         # Analyze multiple files")
		fmt.Println("")
		fmt.Println("Features:")
		fmt.Println("  • Interactive TUI with multiple view modes")
		fmt.Println("  • Performance metrics and Core Web Vitals analysis")
		fmt.Println("  • Multi-file comparison capabilities")
		fmt.Println("  • Professional report export (JSON/CSV/HTML/PDF)")
		fmt.Println("  • Chrome DevTools-style waterfall timeline")
		fmt.Println("  • Advanced filtering and search")
		os.Exit(1)
	}

	// Parse HAR files
	parser := har.NewParser()
	var harFiles []*har.HAR
	
	for _, filepath := range os.Args[1:] {
		harFile, err := parser.ParseFile(filepath)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", filepath, err)
			os.Exit(1)
		}
		
		if err := parser.ValidateHAR(harFile); err != nil {
			fmt.Printf("Invalid HAR file %s: %v\n", filepath, err)
			os.Exit(1)
		}
		
		harFiles = append(harFiles, harFile)
		fmt.Printf("Loaded HAR file: %s (%d entries)\n", filepath, len(harFile.Log.Entries))
	}

	if len(harFiles) == 0 {
		fmt.Println("No valid HAR files found")
		os.Exit(1)
	}

	// Initialize and run TUI
	model := tui.NewModel(harFiles)
	program := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}