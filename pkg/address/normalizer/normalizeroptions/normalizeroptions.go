package normalizeroptions

import "github.com/PortobelloAuth/go-projectusat/pkg/diacritics"

// AddressNormalizationOptions controls exchange/matching variants of normalization.
// Zero value is content form (the same settings used by NewContentNormalizer).
type AddressNormalizationOptions struct {
	// Fuzzy enables FuzzyNormalize* for region and street suffix.
	Fuzzy bool
	// SecondaryAsHash rewrites secondary designators to "#" for matching
	// (not correct for content storage; for exchange/matching only).
	SecondaryAsHash bool
	DiacriticMode   diacritics.DiacriticMode
}

func Content() *AddressNormalizationOptions {
	return &AddressNormalizationOptions{
		Fuzzy:           false,
		SecondaryAsHash: false,
		DiacriticMode:   diacritics.KeepDiacritics,
	}
}

func Matching() *AddressNormalizationOptions {
	return &AddressNormalizationOptions{
		Fuzzy:           true,
		SecondaryAsHash: true,
		DiacriticMode:   diacritics.SubstituteDiacritics,
	}
}
