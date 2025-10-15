package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "version", "--version", "-v":
		fmt.Printf("olsen version %s\n", version)
		fmt.Println("Photo indexer and explorer")
		fmt.Println("Copyright 2025")
		os.Exit(0)
	case "help", "--help", "-h":
		printUsage()
		os.Exit(0)
	case "index":
		handleIndex()
	case "explore":
		handleExplore()
	case "analyze":
		handleAnalyze()
	case "stats":
		handleStats()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Olsen - Photo Indexer and Explorer")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  olsen <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  index      Index photos from a directory (not yet implemented)")
	fmt.Println("  explore    Start web interface to browse photos (not yet implemented)")
	fmt.Println("  analyze    Detect bursts and duplicates (not yet implemented)")
	fmt.Println("  stats      Display database statistics (not yet implemented)")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println("")
	fmt.Println("Run 'olsen <command> --help' for more information on a command.")
	fmt.Println("")
	fmt.Println("Note: This is version 0.1.0-dev. Full CLI implementation is in progress.")
	fmt.Println("      Use shell scripts (./indexphotos.sh, ./explorer.sh) for current functionality.")
}

func handleIndex() {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")
	workers := fs.Int("w", 4, "Number of worker threads")

	fs.Usage = func() {
		fmt.Println("Usage: olsen index <directory> [options]")
		fmt.Println("")
		fmt.Println("Index photos from a directory into a SQLite database.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println("")
		fmt.Println("Status: Not yet implemented. Use ./indexphotos.sh instead.")
	}

	// Parse flags even for help
	if err := fs.Parse(os.Args[2:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	// Check for help flag
	if fs.NFlag() == 0 && fs.NArg() == 0 {
		fs.Usage()
		os.Exit(0)
	}

	// Validate required argument
	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: Photo directory is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	photoDir := fs.Arg(0)

	fmt.Println("Index command is not yet fully implemented.")
	fmt.Printf("  Directory: %s\n", photoDir)
	fmt.Printf("  Database: %s\n", *db)
	fmt.Printf("  Workers: %d\n", *workers)
	fmt.Println("")
	fmt.Println("Please use ./indexphotos.sh for now:")
	fmt.Printf("  ./indexphotos.sh %s --db %s --workers %d\n", photoDir, *db, *workers)
	os.Exit(1)
}

func handleExplore() {
	fs := flag.NewFlagSet("explore", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")
	addr := fs.String("addr", "localhost:8080", "Listen address")
	open := fs.Bool("open", false, "Open browser automatically")

	fs.Usage = func() {
		fmt.Println("Usage: olsen explore [options]")
		fmt.Println("")
		fmt.Println("Start web interface to browse indexed photos.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println("")
		fmt.Println("Status: Not yet implemented. Use ./explorer.sh instead.")
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	fmt.Println("Explore command is not yet fully implemented.")
	fmt.Printf("  Database: %s\n", *db)
	fmt.Printf("  Address: %s\n", *addr)
	fmt.Printf("  Open browser: %v\n", *open)
	fmt.Println("")
	fmt.Println("Please use ./explorer.sh for now:")
	openFlag := ""
	if *open {
		openFlag = " --open"
	}
	fmt.Printf("  ./explorer.sh --db %s --addr %s%s\n", *db, *addr, openFlag)
	os.Exit(1)
}

func handleAnalyze() {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")

	fs.Usage = func() {
		fmt.Println("Usage: olsen analyze [options]")
		fmt.Println("")
		fmt.Println("Detect bursts and duplicates in indexed photos.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println("")
		fmt.Println("Status: Not yet implemented.")
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	fmt.Println("Analyze command is not yet implemented.")
	fmt.Printf("  Database: %s\n", *db)
	os.Exit(1)
}

func handleStats() {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")

	fs.Usage = func() {
		fmt.Println("Usage: olsen stats [options]")
		fmt.Println("")
		fmt.Println("Display statistics about indexed photos.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println("")
		fmt.Println("Status: Not yet implemented.")
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	fmt.Println("Stats command is not yet implemented.")
	fmt.Printf("  Database: %s\n", *db)
	os.Exit(1)
}
