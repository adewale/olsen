#!/bin/bash
# indexphotos.sh - Index photos from a directory into SQLite database

set -e

# Default values
PHOTO_DIR=""
DB_FILE="photos.db"
WORKERS=4

# Usage function
usage() {
    echo "Usage: $0 <photo_directory> [options]"
    echo ""
    echo "Arguments:"
    echo "  photo_directory    Path to directory containing photos (required)"
    echo ""
    echo "Options:"
    echo "  -d, --db FILE      Database file path (default: photos.db)"
    echo "  -w, --workers N    Number of worker threads (default: 4)"
    echo "  -h, --help         Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 ~/Pictures"
    echo "  $0 /mnt/photos --db /var/olsen/photos.db --workers 8"
    exit 1
}

# Parse arguments
if [ $# -eq 0 ]; then
    usage
fi

# Check for help first
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
fi

PHOTO_DIR="$1"
shift

while [ $# -gt 0 ]; do
    case "$1" in
        -d|--db)
            DB_FILE="$2"
            shift 2
            ;;
        -w|--workers)
            WORKERS="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Error: Unknown option: $1"
            usage
            ;;
    esac
done

# Validate photo directory
if [ ! -d "$PHOTO_DIR" ]; then
    echo "Error: Photo directory does not exist: $PHOTO_DIR"
    exit 1
fi

# Build olsen if not present or outdated
if [ ! -f "bin/olsen" ]; then
    echo "Building olsen with CGO support (required for SQLite)..."
    make build-raw
fi

# Create database directory if it doesn't exist
DB_DIR=$(dirname "$DB_FILE")
if [ ! -d "$DB_DIR" ]; then
    mkdir -p "$DB_DIR"
fi

# Run indexer
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔒 READ-ONLY: This indexer will NEVER modify your photo files"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Indexing photos from: $PHOTO_DIR"
echo "Database: $DB_FILE"
echo "Workers: $WORKERS"
echo ""

./bin/olsen index --db "$DB_FILE" --w "$WORKERS" "$PHOTO_DIR"

echo ""
echo "✓ Indexing complete!"
echo "Database saved to: $DB_FILE"
echo ""
echo "Next steps:"
echo "  ./bin/olsen analyze --db $DB_FILE     # Run burst detection"
echo "  ./bin/olsen stats --db $DB_FILE       # View statistics"
echo "  ./explorer.sh --db $DB_FILE           # Browse with web interface"
