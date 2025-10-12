#!/bin/bash

# Script to add realistic EXIF metadata to test images
# This creates more diverse and realistic test data

set -e

# Check if exiftool is available
if ! command -v exiftool &> /dev/null; then
    echo "Error: exiftool is required but not installed"
    echo "Install with: brew install exiftool"
    exit 1
fi

echo "Adding realistic EXIF metadata to test images..."

# Array of realistic camera/lens combinations
cameras=(
    "Canon:EOS R5:RF24-70mm f/2.8 L IS USM"
    "Canon:EOS R6:RF85mm f/1.2 L USM"
    "Nikon:Z9:NIKKOR Z 24-120mm f/4 S"
    "Sony:α7 IV:FE 70-200mm f/2.8 GM OSS II"
    "Fujifilm:X-T5:XF 16-55mm f/2.8 R LM WR"
)

# Counter for varying metadata
counter=0

# Process color test images
if [ -d "testdata/color_test" ]; then
    echo "Processing color_test images..."
    for img in testdata/color_test/*.jpg; do
        [ -f "$img" ] || continue

        # Calculate a unique date (different images throughout 2025)
        month=$((1 + ($counter % 12)))
        day=$((1 + ($counter % 28)))
        hour=$((6 + ($counter % 16)))  # Between 6 AM and 10 PM
        minute=$(($counter % 60))

        # Pick camera/lens combination
        cam_idx=$(($counter % ${#cameras[@]}))
        IFS=':' read -r make model lens <<< "${cameras[$cam_idx]}"

        # Calculate realistic technical params based on time of day
        if [ $hour -lt 10 ] || [ $hour -gt 18 ]; then
            # Low light (morning/evening)
            iso=$((800 + ($counter % 3200)))
            aperture="2.8"
        else
            # Daylight
            iso=$((100 + ($counter % 400)))
            aperture="5.6"
        fi

        # Focal length varies by lens type
        if [[ "$lens" == *"24-70"* ]]; then
            focal=$((24 + ($counter % 47)))
        elif [[ "$lens" == *"70-200"* ]]; then
            focal=$((70 + ($counter % 131)))
        elif [[ "$lens" == *"85"* ]]; then
            focal=85
        else
            focal=$((16 + ($counter % 105)))
        fi

        # Add EXIF metadata
        exiftool -overwrite_original \
            -Make="$make" \
            -Model="$model" \
            -LensModel="$lens" \
            -DateTimeOriginal="2025:$(printf "%02d" $month):$(printf "%02d" $day) $(printf "%02d" $hour):$(printf "%02d" $minute):00" \
            -ISO="$iso" \
            -FNumber="$aperture" \
            -FocalLength="$focal" \
            -ExposureTime="1/$((125 * (1 + ($counter % 8))))" \
            -FocalLengthIn35mmFormat="$focal" \
            "$img" > /dev/null 2>&1

        echo "  ✓ $(basename "$img"): $make $model, $focal mm, f/$aperture, ISO $iso, 2025-$(printf "%02d" $month)-$(printf "%02d" $day) $(printf "%02d" $hour):$(printf "%02d" $minute)"

        counter=$((counter + 1))
    done
fi

# Process burst test images
if [ -d "testdata/burst_test" ]; then
    echo "Processing burst_test images..."
    base_time="2025:03:15 10:30:00"

    for img in testdata/burst_test/*.jpg; do
        [ -f "$img" ] || continue

        # Burst photos are taken in rapid succession
        second=$(($counter % 60))

        exiftool -overwrite_original \
            -Make="Canon" \
            -Model="EOS R5" \
            -LensModel="RF100-500mm f/4.5-7.1 L IS USM" \
            -DateTimeOriginal="$base_time" \
            -SubSecTimeOriginal="$(printf "%03d" $(($counter * 100)))" \
            -ISO="3200" \
            -FNumber="5.6" \
            -FocalLength="400" \
            -ExposureTime="1/1000" \
            -FocalLengthIn35mmFormat="400" \
            "$img" > /dev/null 2>&1

        echo "  ✓ $(basename "$img"): Burst sequence photo"

        counter=$((counter + 1))
    done
fi

# Process DNG test images if they exist
if [ -d "testdata/dng" ]; then
    echo "Processing DNG images..."
    for img in testdata/dng/*.dng; do
        [ -f "$img" ] || continue

        # Add varied metadata for DNG files
        month=$((1 + ($counter % 12)))
        day=$((1 + ($counter % 28)))
        hour=$((8 + ($counter % 12)))

        cam_idx=$(($counter % ${#cameras[@]}))
        IFS=':' read -r make model lens <<< "${cameras[$cam_idx]}"

        exiftool -overwrite_original \
            -Make="$make" \
            -Model="$model" \
            -LensModel="$lens" \
            -DateTimeOriginal="2025:$(printf "%02d" $month):$(printf "%02d" $day) $(printf "%02d" $hour):00:00" \
            -ISO="$((200 + ($counter % 1600)))" \
            -FNumber="$((2 + ($counter % 6))).8" \
            -FocalLength="$((35 + ($counter % 100)))" \
            -FocalLengthIn35mmFormat="$((35 + ($counter % 100)))" \
            "$img" > /dev/null 2>&1

        echo "  ✓ $(basename "$img"): $make $model"

        counter=$((counter + 1))
    done
fi

echo ""
echo "✅ Added realistic EXIF metadata to $counter images"
echo ""
echo "Metadata includes:"
echo "  • Dates: Throughout 2025 with varied times"
echo "  • Cameras: Canon, Nikon, Sony, Fujifilm"
echo "  • Technical settings: Varied ISO, aperture, focal length"
echo "  • Realistic combinations based on lighting conditions"
echo ""
echo "Now re-index the database:"
echo "  rm other_photos.db"
echo "  ./bin/olsen index testdata --db other_photos.db -w 4"
