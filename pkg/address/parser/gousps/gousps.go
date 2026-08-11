package gousps

import (
	"fmt"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/my-eq/go-usps/parser"
)

type GoUSPSParser struct{}

func New() (*GoUSPSParser, error) {
	return &GoUSPSParser{}, nil
}

func (g *GoUSPSParser) Parse(input string) (*address.Address, error) {
	parsed, diagnostics := parser.Parse(input)

	for _, d := range diagnostics {
		fmt.Printf("%s: %s\n", d.Severity, d.Message)
	}

	return &address.Address{
		BusinessName:        parsed.Firm,
		PrimaryNumber:       parsed.HouseNumber,
		Predirectional:      parsed.PreDirectional,
		StreetName:          parsed.StreetName,
		StreetSuffix:        parsed.StreetSuffix,
		Postdirectional:     parsed.PostDirectional,
		SecondaryDesignator: parsed.SecondaryUnit,
		SecondaryNumber:     parsed.SecondaryNumber,
		City:                parsed.City,
		Region:              parsed.State,
		Postal:              parsed.ZIPCode,
	}, nil
}
