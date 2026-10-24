package normalizeroptions

import (
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

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

// AddressNormalizationFnOptions combines the normalizer's AddressNormalizationOptions
// with a copy of the original address to be normalized, so that normalization
// functions can tell what an earlier stage of the normalizer has already changed.
type AddressNormalizationFnOptions struct {
	AddressNormalizationOptions
	// OriginalAddress is the address as it entered normalization. Nil when
	// a caller has none to give.
	OriginalAddress *address.Address
}

func NewFnOptions(a *address.Address, o AddressNormalizationOptions) AddressNormalizationFnOptions {
	return AddressNormalizationFnOptions{
		AddressNormalizationOptions: o,
		OriginalAddress:             a.Clone(),
	}
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
