package normalizer

import (
	"errors"
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer/normalizerfn"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/normalizer/normalizeroptions"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/diacritics"
)

// NormalizingAddressType is an interface describing an AddressType that has a
// Normalize() method because it implements its own Normalization rules. It may
// or may not choose to employ normalization from shared address components, etc.
type NormalizingAddressType interface {
	address.AddressType
	Normalize(a *address.Address, o normalizeroptions.AddressNormalizationOptions) (*address.Address, error)
}

// Address is a Project US@ structured patient address.
// Empty string means unknown / not present.
type Normalizer struct {
	parser  *parser.Parser
	Options normalizeroptions.AddressNormalizationOptions
}

// NewParsingNomalizer creates a Normalizer that is coupled to a Parser. It will check whether
// an address it recieves was parsed by the Parser and, if not, format and reparse it.
func NewParsingNomalizer(p *parser.Parser, opts normalizeroptions.AddressNormalizationOptions) *Normalizer {
	return &Normalizer{
		parser:  p,
		Options: opts,
	}
}

func NewNomalizer(opts normalizeroptions.AddressNormalizationOptions) *Normalizer {
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
	return NewNomalizer(*normalizeroptions.Content())
}

// NewMatchingNomalizer returns a Normalizer with options appropriate for normalizing
// an address for comparison. The Project US@ standard refers to this as
// the "Exchange" use case. The address is returned in uppercase, with standard
// abbreviations. Diacritics are substituted. Numbered Secondary Units use the hash symbol
// to allow for matching addresses where the secondary unit type is not known. Empty
// optional fields stay blank. Unrecognized non-empty controlled vocabulary (region,
// directionals, street suffix, secondary designator) returns an error.
func NewMatchingNomalizer() *Normalizer {
	return NewNomalizer(*normalizeroptions.Matching())
}

// Normalize applies the Normalizer's AddressNormalizationOptions to the Address
func (n *Normalizer) Normalize(a *address.Address) (*address.Address, error) {
	// If our parser didn't parse the address, we'll reparse it to ensure that
	// the address has the correct address type and values in the correct fields.
	if n.parser != nil && !n.parser.IsOurs(a) {
		// If our parser didn't build the address, format it and reparse it
		f, err := a.Format(address.FormatOptions{
			SingleLine:    false,
			DiacriticMode: diacritics.KeepDiacritics,
		})
		if err != nil {
			return nil, err
		}
		a, err = n.parser.Parse(f)
		if err != nil {
			return nil, err
		}
	}

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

	out, err := normalizerfn.NormalizeAddressFn(a, n.Options)
	if err != nil && !errors.Is(err, normalizerfn.Done) {
		return nil, err
	}

	// The type is how the address formats; normalizing the fields does not
	// change which kind of address they make.
	if out.Type != a.Type {
		return nil, fmt.Errorf("normalized address type does not match input type")
	}

	return out, nil
}
