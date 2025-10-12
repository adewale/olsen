#!/usr/bin/env python3
"""
Analyze performance statistics for a specific file type or pattern.

Usage:
    python3 analyze_filetype.py perfstats.json "*.DNG"
    python3 analyze_filetype.py perfstats.json "L10" --pattern
"""

import json
import sys
import argparse


def load_perfstats(filename):
    """Load perfstats JSON file."""
    with open(filename) as f:
        return json.load(f)


def filter_photos(detailed, pattern, use_pattern=False):
    """Filter photos by extension or pattern."""
    if use_pattern:
        return [p for p in detailed if pattern in p['FilePath']]
    else:
        # Treat as file extension
        ext = pattern if pattern.startswith('.') else f'.{pattern}'
        return [p for p in detailed if p['FilePath'].upper().endswith(ext.upper())]


def coefficient_of_variation(values):
    """Calculate coefficient of variation (CV %)."""
    if not values or len(values) < 2:
        return 0
    mean = sum(values) / len(values)
    if mean == 0:
        return 0
    variance = sum((v - mean) ** 2 for v in values) / len(values)
    std_dev = variance ** 0.5
    return (std_dev / mean) * 100


def get_percentile(sorted_values, percentile):
    """Get percentile value from sorted list."""
    if not sorted_values:
        return 0
    idx = int(len(sorted_values) * percentile)
    return sorted_values[min(idx, len(sorted_values) - 1)]


def analyze_photos(photos, label):
    """Analyze a set of photos."""
    if not photos:
        print(f"No photos found matching pattern!")
        return

    n = len(photos)

    print(f"{label.upper()} ANALYSIS ({n} photos)")
    print("=" * 70)
    print()

    # Calculate totals
    total_time = sum(p['TotalTime'] for p in photos)
    hash_time = sum(p['HashTime'] for p in photos)
    metadata_time = sum(p['MetadataTime'] for p in photos)
    decode_time = sum(p['ImageDecodeTime'] for p in photos)
    thumbnail_time = sum(p['ThumbnailTime'] for p in photos)
    color_time = sum(p['ColorTime'] for p in photos)
    phash_time = sum(p['PerceptualHashTime'] for p in photos)
    inference_time = sum(p['InferenceTime'] for p in photos)
    db_time = sum(p['DatabaseTime'] for p in photos)
    total_bytes = sum(p['FileSize'] for p in photos)

    # Calculate averages (convert nanoseconds to milliseconds)
    avg_total = total_time / n / 1_000_000
    avg_hash = hash_time / n / 1_000_000
    avg_metadata = metadata_time / n / 1_000_000
    avg_decode = decode_time / n / 1_000_000
    avg_thumbnail = thumbnail_time / n / 1_000_000
    avg_color = color_time / n / 1_000_000
    avg_phash = phash_time / n / 1_000_000
    avg_inference = inference_time / n / 1_000_000
    avg_db = db_time / n / 1_000_000

    print(f"Count:              {n} photos")
    print(f"Total Size:         {total_bytes/1024/1024:.1f} MB")
    print(f"Avg File Size:      {total_bytes/n/1024/1024:.1f} MB")
    print(f"Throughput:         {(total_bytes/1024/1024) / (total_time/1_000_000_000):.2f} MB/s")
    print()

    print("AVERAGE TIMINGS PER PHOTO:")
    print("-" * 70)
    stages = [
        ('Hash', avg_hash),
        ('Metadata', avg_metadata),
        ('Image Decode', avg_decode),
        ('Thumbnails', avg_thumbnail),
        ('Color Extract', avg_color),
        ('Perceptual Hash', avg_phash),
        ('Inference', avg_inference),
        ('Database', avg_db),
    ]

    print(f"{'Stage':<18} {'Time (ms)':>12} {'% of Total':>12}")
    print("-" * 70)
    for name, time_ms in stages:
        pct = (time_ms / avg_total) * 100 if avg_total > 0 else 0
        print(f"{name:<18} {time_ms:>11.2f}ms {pct:>11.2f}%")

    print(f"{'TOTAL':<18} {avg_total:>11.2f}ms {'100.00%':>12}")
    print()

    print("DISTRIBUTION STATS:")
    print("-" * 70)

    # Get sorted values for percentiles
    decode_times = sorted([p['ImageDecodeTime']/1_000_000 for p in photos])
    thumb_times = sorted([p['ThumbnailTime']/1_000_000 for p in photos])
    color_times = sorted([p['ColorTime']/1_000_000 for p in photos])
    total_times = sorted([p['TotalTime']/1_000_000 for p in photos])

    print(f"{'Stage':<18} {'Min':>10} {'Median':>10} {'P95':>10} {'Max':>10}")
    print("-" * 70)
    print(f"{'Image Decode':<18} {decode_times[0]:>9.0f}ms {get_percentile(decode_times, 0.5):>9.0f}ms {get_percentile(decode_times, 0.95):>9.0f}ms {decode_times[-1]:>9.0f}ms")
    print(f"{'Thumbnails':<18} {thumb_times[0]:>9.0f}ms {get_percentile(thumb_times, 0.5):>9.0f}ms {get_percentile(thumb_times, 0.95):>9.0f}ms {thumb_times[-1]:>9.0f}ms")
    print(f"{'Color Extract':<18} {color_times[0]:>9.0f}ms {get_percentile(color_times, 0.5):>9.0f}ms {get_percentile(color_times, 0.95):>9.0f}ms {color_times[-1]:>9.0f}ms")
    print(f"{'TOTAL':<18} {total_times[0]:>9.0f}ms {get_percentile(total_times, 0.5):>9.0f}ms {get_percentile(total_times, 0.95):>9.0f}ms {total_times[-1]:>9.0f}ms")
    print()

    print("VARIABILITY (Coefficient of Variation %):")
    print("-" * 70)
    variability = [
        ('Hash', coefficient_of_variation([p['HashTime'] for p in photos])),
        ('Metadata', coefficient_of_variation([p['MetadataTime'] for p in photos])),
        ('Image Decode', coefficient_of_variation([p['ImageDecodeTime'] for p in photos])),
        ('Thumbnails', coefficient_of_variation([p['ThumbnailTime'] for p in photos])),
        ('Color Extract', coefficient_of_variation([p['ColorTime'] for p in photos])),
        ('Perceptual Hash', coefficient_of_variation([p['PerceptualHashTime'] for p in photos])),
        ('Database', coefficient_of_variation([p['DatabaseTime'] for p in photos])),
    ]

    for name, cv in sorted(variability, key=lambda x: x[1], reverse=True):
        print(f"{name:<18} {cv:>10.1f}% CV")


def main():
    parser = argparse.ArgumentParser(
        description='Analyze performance for specific file types',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Analyze all DNG files
  python3 analyze_filetype.py perfstats.json ".DNG"

  # Analyze Leica files (pattern matching)
  python3 analyze_filetype.py perfstats.json "L10" --pattern

  # Analyze JPEGs
  python3 analyze_filetype.py perfstats.json ".jpeg"
        """
    )
    parser.add_argument('perfstats_file', help='Path to perfstats JSON file')
    parser.add_argument('pattern', help='File extension (e.g., ".DNG") or pattern to match')
    parser.add_argument('--pattern', '-p', action='store_true',
                       help='Use substring matching instead of extension matching')

    args = parser.parse_args()

    data = load_perfstats(args.perfstats_file)
    photos = filter_photos(data['detailed'], args.pattern, args.pattern)

    label = f"{args.pattern} files" if args.pattern else f"{args.pattern} pattern"
    analyze_photos(photos, label)


if __name__ == '__main__':
    main()
