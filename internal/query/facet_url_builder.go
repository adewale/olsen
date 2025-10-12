package query

import (
	"fmt"
	"strings"
)

// FacetURLBuilder handles URL generation for facet values
// This is separated for independent testing
type FacetURLBuilder struct {
	mapper *URLMapper
}

// NewFacetURLBuilder creates a new facet URL builder
func NewFacetURLBuilder(mapper *URLMapper) *FacetURLBuilder {
	return &FacetURLBuilder{mapper: mapper}
}

// BuildURLsForFacets populates URL fields for all facet values
// Key principle: Facet URLs should REFINE (add to) existing filters, not replace them
func (b *FacetURLBuilder) BuildURLsForFacets(facets *FacetCollection, baseParams QueryParams) {
	if facets.ColourName != nil {
		b.buildColourURLs(facets.ColourName, baseParams)
	}
	if facets.Year != nil {
		b.buildYearURLs(facets.Year, baseParams)
	}
	if facets.Month != nil {
		b.buildMonthURLs(facets.Month, baseParams)
	}
	if facets.Camera != nil {
		b.buildCameraURLs(facets.Camera, baseParams)
	}
	if facets.Lens != nil {
		b.buildLensURLs(facets.Lens, baseParams)
	}
	if facets.TimeOfDay != nil {
		b.buildTimeOfDayURLs(facets.TimeOfDay, baseParams)
	}
	if facets.Season != nil {
		b.buildSeasonURLs(facets.Season, baseParams)
	}
	if facets.FocalCategory != nil {
		b.buildFocalCategoryURLs(facets.FocalCategory, baseParams)
	}
	if facets.ShootingCondition != nil {
		b.buildShootingConditionURLs(facets.ShootingCondition, baseParams)
	}
	if facets.InBurst != nil {
		b.buildBurstURLs(facets.InBurst, baseParams)
	}
}

func (b *FacetURLBuilder) buildColourURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		if facet.Values[i].Selected {
			// Already selected - URL should REMOVE this filter
			p.ColourName = nil
		} else {
			// Not selected - URL should ADD this filter (preserving others)
			p.ColourName = []string{facet.Values[i].Value}
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildYearURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		year := 0
		fmt.Sscanf(facet.Values[i].Value, "%d", &year)

		if facet.Values[i].Selected {
			// Already selected - remove year filter, preserve all other filters
			p.Year = nil
		} else {
			// Add year filter, preserve all other filters
			// The facet computation will determine if this state has results
			p.Year = &year
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildMonthURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		month := 0
		fmt.Sscanf(facet.Values[i].Value, "%d", &month)

		if facet.Values[i].Selected {
			// Already selected - remove month filter, preserve all other filters
			p.Month = nil
		} else {
			// Add month filter, preserve all other filters
			// The facet computation will determine if this state has results
			p.Month = &month
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildCameraURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		parts := strings.SplitN(facet.Values[i].Value, " ", 2)

		if facet.Values[i].Selected {
			// Already selected - remove camera filter
			p.CameraMake = nil
			p.CameraModel = nil
		} else if len(parts) == 2 {
			// Add camera filter (preserving other filters)
			p.CameraMake = []string{parts[0]}
			p.CameraModel = []string{parts[1]}
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildLensURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		if facet.Values[i].Selected {
			// Already selected - remove lens filter
			p.LensModel = nil
		} else {
			// Add lens filter (preserving other filters)
			p.LensModel = []string{facet.Values[i].Value}
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildTimeOfDayURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		if facet.Values[i].Selected {
			// Already selected - remove from list
			p.TimeOfDay = removeFromSlice(p.TimeOfDay, facet.Values[i].Value)
		} else {
			// Add to list (multi-select support)
			p.TimeOfDay = append(p.TimeOfDay, facet.Values[i].Value)
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildSeasonURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		if facet.Values[i].Selected {
			p.Season = removeFromSlice(p.Season, facet.Values[i].Value)
		} else {
			p.Season = append(p.Season, facet.Values[i].Value)
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildFocalCategoryURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		if facet.Values[i].Selected {
			p.FocalCategory = removeFromSlice(p.FocalCategory, facet.Values[i].Value)
		} else {
			p.FocalCategory = append(p.FocalCategory, facet.Values[i].Value)
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildShootingConditionURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		if facet.Values[i].Selected {
			p.ShootingCondition = removeFromSlice(p.ShootingCondition, facet.Values[i].Value)
		} else {
			p.ShootingCondition = append(p.ShootingCondition, facet.Values[i].Value)
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}

func (b *FacetURLBuilder) buildBurstURLs(facet *Facet, baseParams QueryParams) {
	for i := range facet.Values {
		p := baseParams
		if facet.Values[i].Selected {
			// Already selected - remove burst filter
			p.InBurst = nil
		} else {
			// Add burst filter
			inBurst := facet.Values[i].Value == "yes"
			p.InBurst = &inBurst
		}
		facet.Values[i].URL = b.mapper.BuildFullURL(p)
	}
}
