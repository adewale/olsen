#!/usr/bin/env python3
"""
Compare performance statistics between two perfstats JSON files.

Usage:
    python3 compare_datasets.py perfstats_small.json perfstats_large.json
"""

import json
import sys


def load_perfstats(filename):
    """Load perfstats JSON file."""
    with open(filename) as f:
        return json.load(f)


def compare_summaries(small, large):
    """Compare summary statistics between two datasets."""
    small_summary = small['summary']
    large_summary = large['summary']

    print("PERFORMANCE COMPARISON: Small vs Large Dataset")
    print("=" * 70)
    print()

    print(f"Dataset Size:          {small_summary['ProcessedPhotos']:>8} photos  vs  {large_summary['ProcessedPhotos']:>8} photos")
    print(f"Total Data:            {small_summary['TotalBytes']/1024/1024:>8.1f} MB     vs  {large_summary['TotalBytes']/1024/1024:>8.1f} MB")
    print(f"Throughput:            {small_summary['AvgThroughputMBps']:>8.2f} MB/s   vs  {large_summary['AvgThroughputMBps']:>8.2f} MB/s")
    print()

    print("PIPELINE STAGE BREAKDOWN (% of total time):")
    print("-" * 70)
    stages = [
        ('Hash', 'AvgHashMs'),
        ('Metadata', 'AvgMetadataMs'),
        ('Image Decode', 'AvgImageDecodeMs'),
        ('Thumbnails', 'AvgThumbnailMs'),
        ('Color Extract', 'AvgColorMs'),
        ('Perceptual Hash', 'AvgPerceptualHashMs'),
        ('Inference', 'AvgInferenceMs'),
        ('Database', 'AvgDatabaseMs'),
    ]

    print(f"{'Stage':<18} {'Small %':>10} {'Large %':>10} {'Δ':>10}")
    print("-" * 70)
    for name, key in stages:
        small_pct = (small_summary[key] / small_summary['AvgTotalMs']) * 100
        large_pct = (large_summary[key] / large_summary['AvgTotalMs']) * 100
        delta = large_pct - small_pct
        print(f"{name:<18} {small_pct:>9.2f}% {large_pct:>9.2f}% {delta:>+9.2f}%")

    print()
    print("ABSOLUTE TIMINGS (milliseconds per photo):")
    print("-" * 70)
    print(f"{'Stage':<18} {'Small':>10} {'Large':>10} {'Δ':>10} {'% Change':>10}")
    print("-" * 70)
    for name, key in stages:
        small_ms = small_summary[key]
        large_ms = large_summary[key]
        delta = large_ms - small_ms
        if small_ms > 0:
            pct_change = ((large_ms / small_ms) - 1) * 100
        else:
            pct_change = 0
        print(f"{name:<18} {small_ms:>9.2f}ms {large_ms:>9.2f}ms {delta:>+9.2f}ms {pct_change:>+9.1f}%")

    print()
    small_total = small_summary['AvgTotalMs']
    large_total = large_summary['AvgTotalMs']
    delta = large_total - small_total
    pct_change = ((large_total / small_total) - 1) * 100
    print(f"Total per photo:   {small_total:>9.2f}ms {large_total:>9.2f}ms {delta:>+9.2f}ms {pct_change:>+9.1f}%")


def main():
    if len(sys.argv) != 3:
        print("Usage: python3 compare_datasets.py perfstats_small.json perfstats_large.json")
        sys.exit(1)

    small_file = sys.argv[1]
    large_file = sys.argv[2]

    small = load_perfstats(small_file)
    large = load_perfstats(large_file)

    compare_summaries(small, large)


if __name__ == '__main__':
    main()
