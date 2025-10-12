#!/usr/bin/env python3
"""Create test images for each color classification."""

from PIL import Image
import os

# Image dimensions
WIDTH, HEIGHT = 400, 400

# Define colors with RGB values
colors = {
    # Achromatic colors (these are missing)
    'brown_dominant.jpg': (139, 69, 19),      # Brown: hue=25° (orange range), low saturation
    'grey_dominant.jpg': (128, 128, 128),     # Grey: medium lightness, no saturation
    'black_dominant.jpg': (10, 10, 10),       # Black: very low lightness
    'white_dominant.jpg': (245, 245, 245),    # White: very high lightness
}

output_dir = 'testdata/color_test'

for filename, rgb in colors.items():
    filepath = os.path.join(output_dir, filename)

    # Create solid color image
    img = Image.new('RGB', (WIDTH, HEIGHT), rgb)

    # Save as JPEG
    img.save(filepath, 'JPEG', quality=95)
    print(f"Created {filepath} with RGB{rgb}")

print(f"\nCreated {len(colors)} test images")
