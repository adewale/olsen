#!/bin/bash
# explorer.sh - Start the Olsen photo explorer web interface

set -e

# Default values
DB_FILE="photos.db"
ADDR="localhost:8080"
OPEN_BROWSER=false

# Usage function
usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -d, --db FILE      Database file path (default: photos.db)"
    echo "  -a, --addr ADDR    Listen address (default: localhost:8080)"
    echo "  -o, --open         Open browser automatically"
    echo "  -h, --help         Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0"
    echo "  $0 --db /var/olsen/photos.db --addr 0.0.0.0:3000"
    echo "  $0 --db photos.db --open"
    exit 1
}

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        -d|--db)
            DB_FILE="$2"
            shift 2
            ;;
        -a|--addr)
            ADDR="$2"
            shift 2
            ;;
        -o|--open)
            OPEN_BROWSER=true
            shift
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

# Validate database file
if [ ! -f "$DB_FILE" ]; then
    echo "Error: Database file not found: $DB_FILE"
    echo ""
    echo "To create a database, run:"
    echo "  ./indexphotos.sh <photo_directory> --db $DB_FILE"
    exit 1
fi

# Build olsen if not present
if [ ! -f "olsen" ]; then
    echo "Building olsen..."
    go build -o olsen cmd/olsen/main.go
    echo ""
fi

# Start explorer
echo "Starting Olsen Photo Explorer..."
echo "Database: $DB_FILE"
echo "Address: http://$ADDR"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Build command
CMD="./olsen explore --db \"$DB_FILE\" --addr \"$ADDR\""
if [ "$OPEN_BROWSER" = true ]; then
    CMD="$CMD --open"
fi

# Execute
eval $CMD
