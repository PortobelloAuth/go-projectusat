package normalizer

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

var whitespace = regexp.MustCompile(`\s+`)

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

// NormalizingAddressType is an interface describing an AddressType that has a
// Normalize() method because it implements its own Normalization rules. It may
// or may not choose to employ normalization from shared address components, etc.
type NormalizingAddressType interface {
	address.AddressType
	Normalize(a *address.Address, o AddressNormalizationOptions) (*address.Address, error)
}

// Address is a Project US@ structured patient address.
// Empty string means unknown / not present.
type Normalizer struct {
	Options AddressNormalizationOptions
}

func NewNomalizer(opts AddressNormalizationOptions) *Normalizer {
	return &Normalizer{
		Options: opts,
	}
}

// NewContentNomalizer returns a Normalizer with options appropriate for normalizing
// an address as it is being captured. The Project US@ standard refers to this as
// the "Content" use case. The address is returned in uppercase, with standard
// abbreviations. Diacritics are preserved; callers may pre-run diacritics.Substitute
// if needed. Empty optional fields stay blank. Unrecognized non-empty controlled
// vocabulary (region, directionals, street suffix, secondary designator) returns an
// error.
func NewContentNomalizer() *Normalizer {
	return &Normalizer{}
}

// NewMatchingNomalizer returns a Normalizer with options appropriate for normalizing
// an address for comparison. The Project US@ standard refers to this as
// the "Exchange" use case. The address is returned in uppercase, with standard
// abbreviations. Diacritics are substituted. Numbered Secondary Units use the hash symbol
// to allow for matching addresses where the secondary unit type is not known. Empty
// optional fields stay blank. Unrecognized non-empty controlled vocabulary (region,
// directionals, street suffix, secondary designator) returns an error.
func NewMatchingNomalizer() *Normalizer {
	return &Normalizer{
		Options: AddressNormalizationOptions{
			// Fuzzy:           true,  // TODO: decide whether we really want this to be true or false
			SecondaryAsHash: true,
			DiacriticMode:   diacritics.SubstituteDiacritics,
		},
	}
}

// Normalize applies the Normalizer's AddressNormalizationOptions to the Address
func (n *Normalizer) Normalize(a *address.Address) (*address.Address, error) {
	// TODO: how do we detect when the submitted address has the wrong address type?

	// If the address type is an AddressNormalizingType (it implements its own Normalization
	// rules) employ that Normalization instead of the default.
	if normalizing, ok := a.Type.(NormalizingAddressType); ok {
		normalized, err := normalizing.Normalize(a, n.Options)
		if err != nil {
			return nil, err
		}
		if normalized.Type != a.Type {
			return nil, fmt.Errorf("normalized address type does not match input type")
		}
		return normalized, nil
	}

	out, err := normalizeAddressFn(a, n.Options)
	if err != nil && !errors.Is(err, Done) {
		return nil, err
	}

	// The type is how the address formats; normalizing the fields does not
	// change which kind of address they make.
	if out.Type != a.Type {
		return nil, fmt.Errorf("normalized address type does not match input type")
	}

	return out, nil
}
