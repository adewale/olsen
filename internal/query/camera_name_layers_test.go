package query

import (
	"slices"
	"testing"
)

// TestParsePathMultiWordCameraMake checks that camera make and model are read
// verbatim from their own query parameters, with "+" decoded to a space, so a
// multi-word make such as "Leica Camera AG" survives the URL roundtrip.
//
// The facet side of this bug (the URL builder splitting "Leica Camera AG LEICA
// M11 Monochrom" on its first space) is covered against a real database by
// the TestCameraFacet* tests in camera_facet_bug_test.go.
func TestParsePathMultiWordCameraMake(t *testing.T) {
	mapper := NewURLMapper()

	tests := []struct {
		name      string
		url       string
		wantMake  []string
		wantModel []string
	}{
		{
			name:      "multi-word make",
			url:       "/photos?camera_make=Leica+Camera+AG&camera_model=LEICA+M11+Monochrom&limit=100",
			wantMake:  []string{"Leica Camera AG"},
			wantModel: []string{"LEICA M11 Monochrom"},
		},
		{
			// The URL the old builder produced: the parser must not try to
			// repair it, it takes each parameter as given.
			name:      "make split on its first space is taken as given",
			url:       "/photos?camera_make=Leica&camera_model=Camera+AG+LEICA+M11+Monochrom&limit=100",
			wantMake:  []string{"Leica"},
			wantModel: []string{"Camera AG LEICA M11 Monochrom"},
		},
		{
			name:      "single-word make",
			url:       "/photos?camera_make=Canon&camera_model=EOS+R5&limit=100",
			wantMake:  []string{"Canon"},
			wantModel: []string{"EOS R5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := parseTestURL(t, mapper, tt.url)
			if !slices.Equal(params.CameraMake, tt.wantMake) || !slices.Equal(params.CameraModel, tt.wantModel) {
				t.Errorf("CameraMake=%q CameraModel=%q, want %q %q",
					params.CameraMake, params.CameraModel, tt.wantMake, tt.wantModel)
			}
		})
	}
}
