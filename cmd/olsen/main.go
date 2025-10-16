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

	var err error
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
		err = handleIndex()
	case "explore":
		err = handleExplore()
	case "analyze":
		err = handleAnalyze()
	case "stats":
		err = handleStats()
	case "show":
		err = handleShow()
	case "thumbnail":
		err = handleThumbnail()
	case "verify":
		err = handleVerify()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
	fmt.Println("  index      Index photos from a directory")
	fmt.Println("  explore    Start web interface to browse photos")
	fmt.Println("  analyze    Detect bursts and duplicates")
	fmt.Println("  stats      Display database statistics")
	fmt.Println("  show       Show metadata for a specific photo")
	fmt.Println("  thumbnail  Extract thumbnail from a photo")
	fmt.Println("  verify     Verify database integrity")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println("")
	fmt.Println("Run 'olsen <command> --help' for more information on a command.")
}

func handleIndex() error {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")
	workers := fs.Int("w", 4, "Number of worker threads")
	perfstats := fs.Bool("perfstats", false, "Enable performance statistics")

	fs.Usage = func() {
		fmt.Println("Usage: olsen index <directory> [options]")
		fmt.Println("")
		fmt.Println("Index photos from a directory into a SQLite database.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("photo directory is required")
	}

	photoDir := fs.Arg(0)
	return indexCommand(photoDir, *db, *workers, *perfstats)
}

func handleExplore() error {
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
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	return exploreCommand(*db, *addr, *open)
}

func handleAnalyze() error {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")

	fs.Usage = func() {
		fmt.Println("Usage: olsen analyze [options]")
		fmt.Println("")
		fmt.Println("Detect bursts and duplicates in indexed photos.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	return analyzeCommand(*db)
}

func handleStats() error {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")

	fs.Usage = func() {
		fmt.Println("Usage: olsen stats [options]")
		fmt.Println("")
		fmt.Println("Display statistics about indexed photos.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	return statsCommand(*db)
}

func handleShow() error {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")

	fs.Usage = func() {
		fmt.Println("Usage: olsen show <photo-id> [options]")
		fmt.Println("")
		fmt.Println("Show metadata for a specific photo.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("photo ID is required")
	}

	var photoID int
	if _, err := fmt.Sscanf(fs.Arg(0), "%d", &photoID); err != nil {
		return fmt.Errorf("invalid photo ID: %s", fs.Arg(0))
	}

	return showCommand(*db, photoID)
}

func handleThumbnail() error {
	fs := flag.NewFlagSet("thumbnail", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")
	output := fs.String("o", "thumbnail.jpg", "Output file path")
	size := fs.Int("s", 512, "Thumbnail size (64, 256, 512, or 1024)")

	fs.Usage = func() {
		fmt.Println("Usage: olsen thumbnail <photo-id> [options]")
		fmt.Println("")
		fmt.Println("Extract thumbnail from a photo.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("photo ID is required")
	}

	var photoID int
	if _, err := fmt.Sscanf(fs.Arg(0), "%d", &photoID); err != nil {
		return fmt.Errorf("invalid photo ID: %s", fs.Arg(0))
	}

	return thumbnailCommand(*db, photoID, *output, *size)
}

func handleVerify() error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	db := fs.String("db", "photos.db", "Database file path")

	fs.Usage = func() {
		fmt.Println("Usage: olsen verify [options]")
		fmt.Println("")
		fmt.Println("Verify database integrity.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	return verifyCommand(*db)
}
