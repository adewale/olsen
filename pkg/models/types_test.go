package models

import (
	"reflect"
	"testing"
)

func TestParseThumbnailSize(t *testing.T) {
	t.Parallel()

	valid := map[string]ThumbnailSize{
		"64":   ThumbnailTiny,
		"256":  ThumbnailSmall,
		"512":  ThumbnailMedium,
		"1024": ThumbnailLarge,
	}
	for in, want := range valid {
		got, err := ParseThumbnailSize(in)
		if err != nil {
			t.Errorf("ParseThumbnailSize(%q) returned error: %v", in, err)
		}
		if got != want {
			t.Errorf("ParseThumbnailSize(%q) = %q, want %q", in, got, want)
		}
	}

	for _, in := range []string{"", "0", "63", "65", "128", "2048", "abc", " 64", "64 "} {
		if got, err := ParseThumbnailSize(in); err == nil {
			t.Errorf("ParseThumbnailSize(%q) = %q, want error", in, got)
		}
	}
}

func TestThumbnailSizeFallbackOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		size ThumbnailSize
		want []ThumbnailSize
	}{
		{ThumbnailLarge, []ThumbnailSize{ThumbnailLarge, ThumbnailMedium, ThumbnailSmall, ThumbnailTiny}},
		{ThumbnailMedium, []ThumbnailSize{ThumbnailMedium, ThumbnailSmall, ThumbnailTiny}},
		{ThumbnailSmall, []ThumbnailSize{ThumbnailSmall, ThumbnailTiny}},
		{ThumbnailTiny, []ThumbnailSize{ThumbnailTiny}},
	}
	for _, tc := range cases {
		if got := tc.size.FallbackOrder(); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("FallbackOrder(%q) = %v, want %v", tc.size, got, tc.want)
		}
	}
}
